// Web Server - 边缘侧日志管理可视化界面
// 使用 Fastify + Pino 技术栈
const fastify = require('fastify')({
  logger: {
    level: 'info',
    transport: {
      target: 'pino-pretty',
      options: { colorize: true }
    }
  }
});

const Redis = require('ioredis');
const { Pool } = require('pg');
const path = require('path');

// 配置
const config = {
  redis: process.env.REDIS_ADDR || 'localhost:6379',
  pg: {
    host: process.env.PG_HOST || 'localhost',
    port: process.env.PG_PORT || 5432,
    user: process.env.PG_USER || 'postgres',
    password: process.env.PG_PASSWORD || 'Ba0sight',
    database: process.env.PG_DATABASE || 'edge_logs'
  },
  ollama: process.env.OLLAMA_URL || 'http://localhost:11434'
};

// Redis 客户端
const redis = new Redis(config.redis);

// PostgreSQL 连接池
const pgPool = new Pool(config.pg);

// 注册插件
fastify.register(require('@fastify/cors'), { origin: '*' });
fastify.register(require('@fastify/static'), {
  root: path.join(__dirname, 'public'),
  prefix: '/'
});

// ==================== API 路由 ====================

// 1. 获取日志列表（分页、级别过滤）
fastify.get('/api/logs', async (request, reply) => {
  const { page = 1, pageSize = 50, level, service, node } = request.query;
  
  // 从 Redis 读取日志 - 使用 XREVRANGE 从最新开始
  const total = await redis.xlen('logs:stream');
  
  // 计算偏移量
  const offset = (page - 1) * pageSize;
  
  // 使用 XREVRANGE 获取最新的日志
  let msgs = [];
  if (total > 0) {
    msgs = await redis.call('XREVRANGE', 'logs:stream', '+', '-', 'COUNT', parseInt(pageSize) + offset);
  }
  
  // 解析并过滤日志
  const logs = [];
  let skipped = 0;
  for (const msg of msgs) {
    const id = msg[0];
    const fields = msg[1];
    
    // 跳过前面的记录（实现分页）
    if (skipped < offset) {
      skipped++;
      continue;
    }
    
    // fields 是一个数组 ['data', 'jsonstring']
    if (fields && fields.length >= 2) {
      try {
        const entry = JSON.parse(fields[1]);
        
        // 级别过滤
        if (level && levelNumToString(entry.level) !== level) continue;
        
        // 服务过滤
        if (service && entry.svc !== service) continue;
        
        // 节点过滤
        if (node && entry.node !== node) continue;
        
        logs.push({
          id,
          ...entry,
          levelStr: levelNumToString(entry.level),
          timeStr: formatTime(entry.time)
        });
        
        if (logs.length >= parseInt(pageSize)) break;
      } catch (e) {
        // 忽略解析错误
      }
    }
  }
  
  // 获取服务列表
  const services = await getServices();
  const nodes = await getNodes();
  
  return {
    logs,
    total,
    page: parseInt(page),
    pageSize: parseInt(pageSize),
    services,
    nodes
  };
});

// 2. 实时日志流 (SSE)
fastify.get('/api/logs/stream', async (request, reply) => {
  reply.raw.writeHead(200, {
    'Content-Type': 'text/event-stream',
    'Cache-Control': 'no-cache',
    'Connection': 'keep-alive'
  });
  
  let lastId = '$';
  
  const sendLogs = async () => {
    try {
      const result = await redis.call('XREAD', 'BLOCK', 1000, 'STREAMS', 'logs:stream', lastId);
      if (result) {
        // result 格式: [[streamName, [[id, [field, value, ...]], ...]]]
        for (const stream of result) {
          const messages = stream[1];
          for (const msg of messages) {
            const id = msg[0];
            const fields = msg[1];
            if (fields && fields.length >= 2) {
              try {
                const entry = JSON.parse(fields[1]);
                reply.raw.write(`data: ${JSON.stringify({
                  id,
                  ...entry,
                  levelStr: levelNumToString(entry.level),
                  timeStr: formatTime(entry.time)
                })}\n\n`);
              } catch (e) {}
            }
            lastId = id;
          }
        }
      }
    } catch (e) {
      // 忽略错误，继续
    }
  };
  
  // 定期发送心跳
  const heartbeat = setInterval(() => {
    reply.raw.write(': heartbeat\n\n');
  }, 15000);
  
  // 持续发送日志
  const interval = setInterval(sendLogs, 500);
  
  // 清理
  request.raw.on('close', () => {
    clearInterval(interval);
    clearInterval(heartbeat);
  });
});

// 3. 获取统计信息
fastify.get('/api/stats', async (request, reply) => {
  const total = await redis.xlen('logs:stream');
  
  // 获取最近100条日志的统计
  const msgs = await redis.call('XREVRANGE', 'logs:stream', '+', '-', 'COUNT', 100);
  const levelCount = { trace: 0, debug: 0, info: 0, warn: 0, error: 0, fatal: 0 };
  const serviceCount = {};
  
  if (msgs) {
    for (const msg of msgs) {
      const fields = msg[1];
      if (fields && fields.length >= 2) {
        try {
          const entry = JSON.parse(fields[1]);
          const levelStr = levelNumToString(entry.level);
          levelCount[levelStr] = (levelCount[levelStr] || 0) + 1;
          if (entry.svc) {
            serviceCount[entry.svc] = (serviceCount[entry.svc] || 0) + 1;
          }
        } catch (e) {}
      }
    }
  }
  
  return { total, levelCount, serviceCount };
});

// 4. 规则分析
fastify.post('/api/analyze/rules', async (request, reply) => {
  const { count = 50 } = request.body || {};
  
  // 从 Redis 读取日志
  const msgs = await redis.call('XREVRANGE', 'logs:stream', '+', '-', 'COUNT', count);
  const logs = [];
  
  if (msgs) {
    for (const msg of msgs) {
      const fields = msg[1];
      if (fields && fields.length >= 2) {
        try {
          logs.push(JSON.parse(fields[1]));
        } catch (e) {}
      }
    }
  }
  
  // 规则分析
  const results = analyzeWithRules(logs);
  
  return results;
});

// 5. LLM 分析
fastify.post('/api/analyze/llm', async (request, reply) => {
  const { count = 20, model = 'qwen3.5:9b' } = request.body || {};
  
  // 从 Redis 读取日志
  const msgs = await redis.call('XREVRANGE', 'logs:stream', '+', '-', 'COUNT', count);
  const logs = [];
  
  if (msgs) {
    for (const msg of msgs) {
      const fields = msg[1];
      if (fields && fields.length >= 2) {
        try {
          logs.push(JSON.parse(fields[1]));
        } catch (e) {}
      }
    }
  }
  
  if (logs.length === 0) {
    return { error: '没有日志可分析' };
  }
  
  // 调用 Ollama
  const analysis = await callOllama(logs, model);
  
  return analysis;
});

// 6. 获取分析历史
fastify.get('/api/history', async (request, reply) => {
  const { hours = 24 } = request.query;
  
  const result = await pgPool.query(`
    SELECT id, analysis_time, log_count, match_count, severity_distribution, type_distribution
    FROM analysis_results
    WHERE analysis_time >= NOW() - INTERVAL '1 hour' * $1
    ORDER BY analysis_time DESC
    LIMIT 50
  `, [hours]);
  
  return result.rows;
});

// 7. 获取规则列表
fastify.get('/api/rules', async (request, reply) => {
  return getDefaultRules();
});

// ==================== 辅助函数 ====================

function levelNumToString(level) {
  const levels = { 10: 'trace', 20: 'debug', 30: 'info', 40: 'warn', 50: 'error', 60: 'fatal' };
  if (typeof level === 'number') return levels[level] || 'info';
  if (typeof level === 'string') return level.toLowerCase();
  return 'info';
}

function formatTime(time) {
  if (!time) return new Date().toISOString();
  if (typeof time === 'number') {
    // 毫秒时间戳
    return new Date(time).toISOString();
  }
  if (typeof time === 'string') {
    // ISO 字符串
    return time;
  }
  return String(time);
}

async function getServices() {
  const msgs = await redis.call('XREVRANGE', 'logs:stream', '+', '-', 'COUNT', 100);
  const services = new Set();
  if (msgs) {
    for (const msg of msgs) {
      const fields = msg[1];
      if (fields && fields.length >= 2) {
        try {
          const entry = JSON.parse(fields[1]);
          if (entry.svc) services.add(entry.svc);
        } catch (e) {}
      }
    }
  }
  return Array.from(services);
}

async function getNodes() {
  const msgs = await redis.call('XREVRANGE', 'logs:stream', '+', '-', 'COUNT', 100);
  const nodes = new Set();
  if (msgs) {
    for (const msg of msgs) {
      const fields = msg[1];
      if (fields && fields.length >= 2) {
        try {
          const entry = JSON.parse(fields[1]);
          if (entry.node) nodes.add(entry.node);
        } catch (e) {}
      }
    }
  }
  return Array.from(nodes);
}

function getDefaultRules() {
  return [
    { id: 'PLC-001', name: 'PLC连接失败', severity: 'critical', type: 'error', description: '检测西门子PLC连接超时或失败' },
    { id: 'PLC-002', name: 'PLC地址超出范围', severity: 'high', type: 'error', description: '检测PLC数据块地址超出范围错误' },
    { id: 'PLC-003', name: 'PLC读取失败', severity: 'high', type: 'error', description: '检测PLC数据读取失败' },
    { id: 'PLC-004', name: 'PLC写入延迟', severity: 'medium', type: 'warning', description: '检测PLC写入操作延迟过高' },
    { id: 'TAG-001', name: 'Tag读取失败', severity: 'high', type: 'error', description: '检测SmartTpc Tag值读取失败' },
    { id: 'TAG-002', name: '设备状态异常', severity: 'medium', type: 'warning', description: '检测SmartTpc设备处于非远程/自动状态' },
    { id: 'TAG-003', name: 'PLC连接超时', severity: 'critical', type: 'error', description: '检测SmartTpc PLC连接超时' },
    { id: 'UNKNOWN-001', name: '未知错误类型', severity: 'high', type: 'error', description: '检测未知的错误类型，需要大模型进一步分析' }
  ];
}

function analyzeWithRules(logs) {
  const rules = getDefaultRules();
  const matches = [];
  const severityCount = { critical: 0, high: 0, medium: 0, low: 0 };
  const typeCount = { error: 0, warning: 0, info: 0, anomaly: 0 };
  
  for (const log of logs) {
    const levelStr = levelNumToString(log.level);
    const msg = log.msg || '';
    
    // PLC 规则
    if (levelStr === 'error') {
      if (msg.includes('Connect Failed') || msg.includes('连接失败')) {
        matches.push({ rule: rules[0], log, suggestion: '检查PLC网络连接、设备电源和通讯参数配置' });
        severityCount.critical++;
        typeCount.error++;
      } else if (msg.includes('Address out of range')) {
        matches.push({ rule: rules[1], log, suggestion: '检查数据块地址配置是否正确' });
        severityCount.high++;
        typeCount.error++;
      } else if (log.fields?.error_type === 'read_failed') {
        matches.push({ rule: rules[2], log, suggestion: '检查PLC设备状态、通讯电缆和网络稳定性' });
        severityCount.high++;
        typeCount.error++;
      } else if (msg.includes('读tag值') && msg.includes('失败')) {
        matches.push({ rule: rules[4], log, suggestion: '检查Tag配置和PLC通讯状态' });
        severityCount.high++;
        typeCount.error++;
      } else if (log.fields?.error_type === 'connection_timeout') {
        matches.push({ rule: rules[6], log, suggestion: '检查PLC网络连接和设备电源状态' });
        severityCount.critical++;
        typeCount.error++;
      } else {
        matches.push({ rule: rules[7], log, suggestion: '需要详细分析以确定根本原因', needsLLM: true });
        severityCount.high++;
        typeCount.error++;
      }
    } else if (levelStr === 'warn') {
      if (log.fields?.error_type === 'device_status_error') {
        matches.push({ rule: rules[5], log, suggestion: '检查SmartTpc设备的远程/自动切换状态' });
        severityCount.medium++;
        typeCount.warning++;
      }
    }
  }
  
  return {
    totalLogs: logs.length,
    matchCount: matches.length,
    severityCount,
    typeCount,
    matches: matches.slice(0, 20), // 最多返回20条
    needsLLM: matches.filter(m => m.needsLLM).length
  };
}

async function callOllama(logs, model) {
  const logTexts = logs.slice(0, 10).map((log, i) => {
    return `日志${i + 1}: 服务=${log.svc}, 级别=${levelNumToString(log.level)}, 消息=${log.msg}`;
  });
  
  const prompt = `你是边缘设备运维专家。请用中文分析以下工业控制系统日志并提供诊断建议。

【日志内容】
${logTexts.join('\n')}

【输出要求】
请严格使用中文输出。按以下格式输出：

## 问题诊断
（描述发现的问题）

## 根本原因
（分析问题的根本原因）

## 解决建议
（提供具体的解决措施）`;

  try {
    const response = await fetch(`${config.ollama}/api/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        model,
        prompt,
        stream: false,
        options: { temperature: 0.3, num_predict: 2048 }
      })
    });
    
    const data = await response.json();
    
    // 处理 qwen3.5 的 thinking 字段
    let content = data.response || data.thinking || '';
    
    // 尝试提取中文部分
    if (content && !content.includes('问题诊断')) {
      const idx = content.indexOf('##');
      if (idx !== -1) {
        content = content.substring(idx);
      }
    }
    
    return {
      model,
      content,
      logCount: logs.length,
      timestamp: new Date().toISOString()
    };
  } catch (error) {
    return { error: `调用大模型失败: ${error.message}` };
  }
}

// ==================== 启动服务 ====================

const start = async () => {
  try {
    // 测试 Redis 连接
    await redis.ping();
    fastify.log.info('Redis 连接成功');
    
    // 测试 PostgreSQL 连接
    await pgPool.query('SELECT 1');
    fastify.log.info('PostgreSQL 连接成功');
    
    // 启动服务
    await fastify.listen({ port: 3000, host: '0.0.0.0' });
    fastify.log.info(`Web 服务启动: http://localhost:3000`);
  } catch (err) {
    fastify.log.error(err);
    process.exit(1);
  }
};

start();