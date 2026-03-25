// file-collector - 日志文件采集器
// 用于采集第三方程序的日志文件并接入到边缘日志系统
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/redis/go-redis/v9"
)

// Config 采集器配置
type Config struct {
	RedisAddr   string
	NodeID      string
	ServiceName string
	LogFiles    []string
	Format      string // log4j, logback, syslog, json, auto
}

// LogParser 日志解析器接口
type LogParser interface {
	Parse(line string) (map[string]interface{}, bool)
	Name() string
}

// AutoParser 自动检测格式的解析器
type AutoParser struct {
	parsers []LogParser
}

func NewAutoParser() *AutoParser {
	return &AutoParser{
		parsers: []LogParser{
			NewLog4jParser(),
			NewJSONParser(),
			NewSyslogParser(),
			NewSimpleParser(),
		},
	}
}

func (p *AutoParser) Parse(line string) (map[string]interface{}, bool) {
	for _, parser := range p.parsers {
		if result, ok := parser.Parse(line); ok {
			result["_parser"] = parser.Name()
			return result, true
		}
	}
	return nil, false
}

func (p *AutoParser) Name() string {
	return "auto"
}

// Log4jParser Log4j/Logback格式解析器
// 格式: [LEVEL] 2024-01-15 10:30:45,123 [thread-name] className - message
type Log4jParser struct {
	pattern *regexp.Regexp
}

func NewLog4jParser() *Log4jParser {
	// 支持多种Log4j格式
	pattern := regexp.MustCompile(`^\[(\w+)\]\s+(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}[,.]\d+)\s+\[([^\]]*)\]\s+([^\s]+)\s+-\s+(.*)$`)
	return &Log4jParser{pattern: pattern}
}

func (p *Log4jParser) Parse(line string) (map[string]interface{}, bool) {
	matches := p.pattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, false
	}

	level := strings.ToLower(matches[1])
	// Log4j级别映射到Pino级别
	levelMap := map[string]int{
		"trace": 10, "debug": 20, "info": 30,
		"warn": 40, "warning": 40, "error": 50, "fatal": 60,
	}

	return map[string]interface{}{
		"level":    levelMap[level],
		"levelStr": level,
		"time":     matches[2],
		"thread":   matches[3],
		"class":    matches[4],
		"msg":      matches[5],
		"raw_log":  line,
		"svc":      matches[4], // 使用类名作为服务名
	}, true
}

func (p *Log4jParser) Name() string {
	return "log4j"
}

// JSONParser JSON格式日志解析器
type JSONParser struct{}

func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

func (p *JSONParser) Parse(line string) (map[string]interface{}, bool) {
	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] != '{' {
		return nil, false
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(line), &result); err != nil {
		return nil, false
	}

	// 确保有必要字段
	if _, ok := result["msg"]; !ok {
		return nil, false
	}

	// 处理时间字段
	if t, ok := result["time"].(string); ok {
		result["time"] = t
	} else {
		result["time"] = time.Now().UTC().Format(time.RFC3339)
	}

	// 处理级别字段
	if level, ok := result["level"]; ok {
		switch v := level.(type) {
		case float64:
			result["level"] = int(v)
		case string:
			levelMap := map[string]int{
				"trace": 10, "debug": 20, "info": 30,
				"warn": 40, "warning": 40, "error": 50, "fatal": 60,
			}
			result["level"] = levelMap[strings.ToLower(v)]
		}
	} else {
		result["level"] = 30 // 默认INFO
	}

	result["raw_log"] = line
	return result, true
}

func (p *JSONParser) Name() string {
	return "json"
}

// SyslogParser Syslog格式解析器
// 格式: Jan 15 10:30:45 hostname process[pid]: message
type SyslogParser struct {
	pattern *regexp.Regexp
}

func NewSyslogParser() *SyslogParser {
	pattern := regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+(\S+)\s+(\S+?)(?:\[(\d+)\])?:\s*(.*)$`)
	return &SyslogParser{pattern: pattern}
}

func (p *SyslogParser) Parse(line string) (map[string]interface{}, bool) {
	matches := p.pattern.FindStringSubmatch(line)
	if matches == nil {
		return nil, false
	}

	return map[string]interface{}{
		"level":    30, // 默认INFO
		"time":     matches[1],
		"hostname": matches[2],
		"process":  matches[3],
		"pid":      matches[4],
		"msg":      matches[5],
		"raw_log":  line,
		"svc":      matches[3],
	}, true
}

func (p *SyslogParser) Name() string {
	return "syslog"
}

// SimpleParser 简单格式解析器（用于无法识别的日志）
type SimpleParser struct{}

func NewSimpleParser() *SimpleParser {
	return &SimpleParser{}
}

func (p *SimpleParser) Parse(line string) (map[string]interface{}, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, false
	}

	// 尝试从内容推断级别
	level := 30 // 默认INFO
	lineLower := strings.ToLower(line)
	if strings.Contains(lineLower, "error") || strings.Contains(lineLower, "exception") {
		level = 50
	} else if strings.Contains(lineLower, "warn") {
		level = 40
	} else if strings.Contains(lineLower, "debug") {
		level = 20
	}

	return map[string]interface{}{
		"level":   level,
		"time":    time.Now().UTC().Format(time.RFC3339),
		"msg":     line,
		"raw_log": line,
	}, true
}

func (p *SimpleParser) Name() string {
	return "simple"
}

// FileCollector 文件采集器
type FileCollector struct {
	config    Config
	parser    LogParser
	redis     *redis.Client
	positions map[string]int64 // 文件读取位置
}

// NewFileCollector 创建文件采集器
func NewFileCollector(config Config) (*FileCollector, error) {
	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		PoolSize: 5,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	// 创建解析器
	var parser LogParser
	switch config.Format {
	case "log4j", "logback":
		parser = NewLog4jParser()
	case "json":
		parser = NewJSONParser()
	case "syslog":
		parser = NewSyslogParser()
	case "simple":
		parser = NewSimpleParser()
	default:
		parser = NewAutoParser()
	}

	return &FileCollector{
		config:    config,
		parser:    parser,
		redis:     rdb,
		positions: make(map[string]int64),
	}, nil
}

// Start 启动采集器
func (fc *FileCollector) Start(ctx context.Context) error {
	// 创建文件监控器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("创建文件监控器失败: %w", err)
	}
	defer watcher.Close()

	// 监控日志文件
	watchedFiles := make(map[string]bool)
	for _, file := range fc.config.LogFiles {
		// 检查文件是否存在
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("警告: 日志文件不存在: %s\n", file)
			continue
		}

		// 监控文件所在目录（用于检测文件轮转）
		dir := filepath.Dir(file)
		if !watchedFiles[dir] {
			if err := watcher.Add(dir); err != nil {
				fmt.Printf("警告: 无法监控目录 %s: %v\n", dir, err)
			} else {
				watchedFiles[dir] = true
				fmt.Printf("监控目录: %s\n", dir)
			}
		}

		// 初始读取文件末尾
		fc.positions[file] = fc.getFileSize(file)
		fmt.Printf("开始监控日志文件: %s (位置: %d)\n", file, fc.positions[file])
	}

	// 启动定时扫描goroutine
	scanTicker := time.NewTicker(500 * time.Millisecond)
	defer scanTicker.Stop()

	fmt.Println("\n日志文件采集器运行中... (按 Ctrl+C 退出)")

	for {
		select {
		case <-ctx.Done():
			return nil

		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			// 处理文件事件
			fc.handleFileEvent(ctx, event)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "监控错误: %v\n", err)

		case <-scanTicker.C:
			// 定期检查文件变化
			fc.scanFiles(ctx)
		}
	}
}

// handleFileEvent 处理文件事件
func (fc *FileCollector) handleFileEvent(ctx context.Context, event fsnotify.Event) {
	for _, file := range fc.config.LogFiles {
		if filepath.Base(event.Name) != filepath.Base(file) {
			continue
		}

		switch {
		case event.Op&fsnotify.Write == fsnotify.Write:
			// 文件被写入，读取新内容
			fc.readNewLines(ctx, file)

		case event.Op&fsnotify.Create == fsnotify.Create:
			// 文件被创建（可能是日志轮转）
			if event.Name == file {
				fc.positions[file] = 0
				fmt.Printf("检测到日志文件创建/轮转: %s\n", file)
			}

		case event.Op&fsnotify.Rename == fsnotify.Rename:
			// 文件重命名（可能是日志轮转）
			fmt.Printf("检测到日志文件重命名: %s\n", event.Name)
		}
	}
}

// scanFiles 扫描文件变化
func (fc *FileCollector) scanFiles(ctx context.Context) {
	for _, file := range fc.config.LogFiles {
		currentSize := fc.getFileSize(file)
		lastPos := fc.positions[file]

		if currentSize > lastPos {
			fc.readNewLines(ctx, file)
		} else if currentSize < lastPos {
			// 文件变小了，可能是日志轮转
			fc.positions[file] = 0
			fc.readNewLines(ctx, file)
		}
	}
}

// readNewLines 读取文件新内容
func (fc *FileCollector) readNewLines(ctx context.Context, file string) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	// 定位到上次读取位置
	pos := fc.positions[file]
	_, err = f.Seek(pos, io.SeekStart)
	if err != nil {
		return
	}

	// 读取新行
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024) // 支持大行

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// 解析日志行
		if entry, ok := fc.parser.Parse(line); ok {
			fc.sendToRedis(ctx, entry)
		}
	}

	// 更新读取位置
	newPos, _ := f.Seek(0, io.SeekCurrent)
	fc.positions[file] = newPos
}

// getFileSize 获取文件大小
func (fc *FileCollector) getFileSize(file string) int64 {
	info, err := os.Stat(file)
	if err != nil {
		return 0
	}
	return info.Size()
}

// sendToRedis 发送到Redis
func (fc *FileCollector) sendToRedis(ctx context.Context, entry map[string]interface{}) {
	// 补充必要字段
	if _, ok := entry["svc"]; !ok {
		entry["svc"] = fc.config.ServiceName
	}
	if _, ok := entry["node"]; !ok {
		entry["node"] = fc.config.NodeID
	}
	if _, ok := entry["name"]; !ok {
		entry["name"] = fc.config.ServiceName
	}
	if _, ok := entry["hostname"]; !ok {
		entry["hostname"] = fc.config.NodeID
	}

	// 添加采集时间
	entry["collected_at"] = time.Now().UTC().Format(time.RFC3339)

	// 序列化
	data, _ := json.Marshal(entry)

	// 发送到Redis Stream
	fc.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: "logs:stream",
		MaxLen: 10000,
		Values: map[string]interface{}{
			"data": string(data),
		},
	})
}

// Close 关闭采集器
func (fc *FileCollector) Close() error {
	return fc.redis.Close()
}

func main() {
	var (
		redisAddr   = flag.String("redis", "localhost:6379", "Redis地址")
		nodeID      = flag.String("node", "edge-collector", "节点ID")
		serviceName = flag.String("service", "external-service", "服务名称")
		format      = flag.String("format", "auto", "日志格式: auto/log4j/json/syslog/simple")
		files       = flag.String("files", "", "日志文件路径，多个用逗号分隔")
	)
	flag.Parse()

	if *files == "" {
		fmt.Println("请指定要监控的日志文件 (-files 参数)")
		fmt.Println("\n示例:")
		fmt.Println("  file-collector -files /var/log/app.log")
		fmt.Println("  file-collector -files \"/var/log/app1.log,/var/log/app2.log\"")
		fmt.Println("  file-collector -files logs/app.log -format log4j")
		os.Exit(1)
	}

	// 解析文件列表
	logFiles := strings.Split(*files, ",")
	for i, f := range logFiles {
		logFiles[i] = strings.TrimSpace(f)
	}

	config := Config{
		RedisAddr:   *redisAddr,
		NodeID:      *nodeID,
		ServiceName: *serviceName,
		LogFiles:    logFiles,
		Format:      *format,
	}

	// 创建采集器
	collector, err := NewFileCollector(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建采集器失败: %v\n", err)
		os.Exit(1)
	}
	defer collector.Close()

	// 处理信号
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n收到停止信号，正在退出...")
		cancel()
	}()

	// 启动采集
	fmt.Printf("日志文件采集器启动\n")
	fmt.Printf("Redis: %s\n", config.RedisAddr)
	fmt.Printf("节点: %s\n", config.NodeID)
	fmt.Printf("服务: %s\n", config.ServiceName)
	fmt.Printf("格式: %s\n", config.Format)
	fmt.Printf("监控文件: %v\n", config.LogFiles)

	if err := collector.Start(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "采集器运行错误: %v\n", err)
		os.Exit(1)
	}
}
