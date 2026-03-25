-- 边缘侧日志管理系统 - PostgreSQL 数据库初始化脚本

-- 创建数据库（如果不存在）
-- CREATE DATABASE edge_logs;

-- 分析结果表
CREATE TABLE IF NOT EXISTS analysis_results (
    id SERIAL PRIMARY KEY,
    analysis_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    log_count INTEGER NOT NULL,
    match_count INTEGER NOT NULL,
    severity_distribution JSONB,
    type_distribution JSONB,
    rules_matched JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 规则匹配详情表
CREATE TABLE IF NOT EXISTS rule_matches (
    id SERIAL PRIMARY KEY,
    result_id INTEGER REFERENCES analysis_results(id) ON DELETE CASCADE,
    rule_id VARCHAR(50) NOT NULL,
    rule_name VARCHAR(200) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    rule_type VARCHAR(50) NOT NULL,
    suggestion TEXT,
    log_message TEXT,
    log_service VARCHAR(100),
    log_node VARCHAR(100),
    log_time TIMESTAMP,
    log_fields JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 大模型分析结果表
CREATE TABLE IF NOT EXISTS llm_analysis (
    id SERIAL PRIMARY KEY,
    result_id INTEGER REFERENCES analysis_results(id) ON DELETE CASCADE,
    model_name VARCHAR(100) NOT NULL,
    analysis_content TEXT NOT NULL,
    log_count INTEGER NOT NULL,
    analysis_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 统计数据表 (按小时聚合)
CREATE TABLE IF NOT EXISTS hourly_stats (
    id SERIAL PRIMARY KEY,
    stat_hour TIMESTAMP NOT NULL,
    log_count INTEGER NOT NULL DEFAULT 0,
    error_count INTEGER NOT NULL DEFAULT 0,
    warn_count INTEGER NOT NULL DEFAULT 0,
    info_count INTEGER NOT NULL DEFAULT 0,
    service_stats JSONB,
    node_stats JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(stat_hour)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_analysis_results_time ON analysis_results(analysis_time);
CREATE INDEX IF NOT EXISTS idx_rule_matches_result_id ON rule_matches(result_id);
CREATE INDEX IF NOT EXISTS idx_rule_matches_rule_id ON rule_matches(rule_id);
CREATE INDEX IF NOT EXISTS idx_rule_matches_severity ON rule_matches(severity);
CREATE INDEX IF NOT EXISTS idx_rule_matches_time ON rule_matches(log_time);
CREATE INDEX IF NOT EXISTS idx_llm_analysis_result_id ON llm_analysis(result_id);
CREATE INDEX IF NOT EXISTS idx_llm_analysis_time ON llm_analysis(analysis_time);
CREATE INDEX IF NOT EXISTS idx_hourly_stats_hour ON hourly_stats(stat_hour);

-- 注释
COMMENT ON TABLE analysis_results IS '分析结果主表';
COMMENT ON TABLE rule_matches IS '规则匹配详情表';
COMMENT ON TABLE llm_analysis IS '大模型分析结果表';
COMMENT ON TABLE hourly_stats IS '按小时聚合的统计表';

COMMENT ON COLUMN analysis_results.severity_distribution IS '严重级别分布统计';
COMMENT ON COLUMN analysis_results.type_distribution IS '类型分布统计';
COMMENT ON COLUMN analysis_results.rules_matched IS '匹配的规则ID列表';

COMMENT ON COLUMN rule_matches.log_fields IS '原始日志字段';
COMMENT ON COLUMN hourly_stats.service_stats IS '服务分布统计';
COMMENT ON COLUMN hourly_stats.node_stats IS '节点分布统计';