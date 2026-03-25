// Package logger 提供极简的边缘设备日志记录功能
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Bunyan 日志级别常量
const (
	LevelTrace = 10
	LevelDebug = 20
	LevelInfo  = 30
	LevelWarn  = 40
	LevelError = 50
	LevelFatal = 60
)

// LogEntry 结构化日志条目 (兼容 Bunyan 格式)
type LogEntry struct {
	Name     string                 `json:"name"`     // 服务名称
	Hostname string                 `json:"hostname"` // 节点 ID
	Pid      int                    `json:"pid"`      // 进程 ID
	Level    int                    `json:"level"`    // Bunyan 日志级别
	Msg      string                 `json:"msg"`      // 消息
	Time     string                 `json:"time"`     // ISO 时间
	Svc      string                 `json:"svc"`      // 服务名
	Node     string                 `json:"node"`     // 节点 ID
	Fields   map[string]interface{} `json:"fields,omitempty"`
}

// SimpleLogger 极简日志器
type SimpleLogger struct {
	redis     *redis.Client
	stream    string
	nodeID    string
	service   string
	localFile *os.File
	mu        sync.Mutex
	redisDown bool // Redis 不可用时为 true
}

// NewSimpleLogger 创建日志器
// Redis 连接失败时自动降级，仅使用本地文件日志
func NewSimpleLogger(redisAddr, nodeID, service string) (*SimpleLogger, error) {
	// 本地文件作为备份（跨平台）
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败：%w", err)
	}
	localPath := fmt.Sprintf("%s/%s-%s.log", logsDir, service, nodeID)
	f, err := os.OpenFile(localPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("创建本地日志文件失败：%w", err)
	}

	// 尝试连接 Redis（失败不阻塞，降级为本地日志）
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		PoolSize: 5,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "[WARN] Redis 连接失败 (%s)，仅使用本地文件日志：%v\n", redisAddr, err)
		return &SimpleLogger{
			redis:     rdb,
			stream:    "logs:stream",
			nodeID:    nodeID,
			service:   service,
			localFile: f,
			redisDown: true,
		}, nil
	}

	return &SimpleLogger{
		redis:     rdb,
		stream:    "logs:stream",
		nodeID:    nodeID,
		service:   service,
		localFile: f,
	}, nil
}

// levelToBunyan 转换日志级别到 Bunyan 格式
func levelToBunyan(level string) int {
	switch level {
	case "trace":
		return LevelTrace
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// Log 记录日志 - 同时写入 Redis 和本地
func (l *SimpleLogger) Log(ctx context.Context, level, msg string, fields map[string]interface{}) {
	entry := LogEntry{
		Name:     l.service,
		Hostname: l.nodeID,
		Pid:      os.Getpid(),
		Level:    levelToBunyan(level),
		Msg:      msg,
		Time:     time.Now().UTC().Format(time.RFC3339),
		Svc:      l.service,
		Node:     l.nodeID,
		Fields:   fields,
	}

	data, _ := json.Marshal(entry)

	// 1. 写入 Redis (异步，失败不阻塞)
	if !l.redisDown {
		go func() {
			l.redis.XAdd(ctx, &redis.XAddArgs{
				Stream: l.stream,
				MaxLen: 10000, // 保留最近 1 万条
				Values: map[string]interface{}{
					"data": string(data),
				},
			})
		}()
	}

	// 2. 写入本地文件 (同步，确保不丢)
	l.mu.Lock()
	l.localFile.Write(append(data, '\n'))
	l.mu.Unlock()
}

// Info 信息级别日志
func (l *SimpleLogger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	l.Log(ctx, "info", msg, fields)
}

// Error 错误级别日志
func (l *SimpleLogger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	l.Log(ctx, "error", msg, fields)
}

// Warn 警告级别日志
func (l *SimpleLogger) Warn(ctx context.Context, msg string, fields map[string]interface{}) {
	l.Log(ctx, "warn", msg, fields)
}

// Debug 调试级别日志
func (l *SimpleLogger) Debug(ctx context.Context, msg string, fields map[string]interface{}) {
	l.Log(ctx, "debug", msg, fields)
}

// Close 关闭日志器
func (l *SimpleLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.localFile != nil {
		l.localFile.Close()
	}
	return l.redis.Close()
}
