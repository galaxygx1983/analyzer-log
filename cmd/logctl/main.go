// logctl - 边缘设备日志管理命令行工具
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"edge-log-demo/pkg/logger"
	"edge-log-demo/pkg/rules"
	"edge-log-demo/pkg/storage"

	"github.com/redis/go-redis/v9"
)

func init() {
	// Windows下设置控制台UTF-8编码
	if runtime.GOOS == "windows" {
		exec.Command("cmd", "/c", "chcp", "65001").Run()
	}
}

// 全局PostgreSQL存储
var pgStorage *storage.PostgresStorage

func main() {
	var (
		redisAddr  = flag.String("redis", "localhost:6379", "Redis地址")
		action     = flag.String("action", "help", "操作: tail/follow/rules-analyze/llm-analyze/rules/stats/history/stream/pino")
		count      = flag.Int("count", 20, "显示日志数量")
		level      = flag.String("level", "", "过滤日志级别 (info/warn/error)")
		service    = flag.String("service", "", "过滤服务名称")
		nodeID     = flag.String("node", "", "过滤节点ID")
		ollamaURL  = flag.String("ollama", "http://localhost:11434", "Ollama服务地址")
		model      = flag.String("model", "minimax-m2.7:cloud", "Ollama模型名称")
		watch      = flag.Bool("watch", false, "持续监控模式")
		persist    = flag.Bool("persist", false, "持久化分析结果到PostgreSQL")
		pgHost     = flag.String("pg-host", "localhost", "PostgreSQL主机")
		pgPort     = flag.Int("pg-port", 5432, "PostgreSQL端口")
		pgUser     = flag.String("pg-user", "postgres", "PostgreSQL用户")
		pgPassword = flag.String("pg-password", "Ba0sight", "PostgreSQL密码")
		pgDatabase = flag.String("pg-database", "edge_logs", "PostgreSQL数据库")
		hours      = flag.Int("hours", 24, "查询历史的小时数")
		ruleFilter = flag.String("rule", "", "过滤规则ID")
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化PostgreSQL存储（如果需要持久化）
	if *persist || *action == "history" || *action == "rule-stats" {
		var err error
		pgStorage, err = storage.NewPostgresStorage(storage.PostgresConfig{
			Host:     *pgHost,
			Port:     *pgPort,
			User:     *pgUser,
			Password: *pgPassword,
			Database: *pgDatabase,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "连接PostgreSQL失败: %v\n", err)
			os.Exit(1)
		}
		defer pgStorage.Close()

		// 初始化表结构
		if err := pgStorage.InitTables(); err != nil {
			fmt.Fprintf(os.Stderr, "初始化表结构失败: %v\n", err)
			os.Exit(1)
		}
	}

	// 处理信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	switch *action {
	case "tail":
		tailLogs(ctx, *redisAddr, *count, *level, *service, *nodeID)

	case "follow":
		followLogs(ctx, *redisAddr, *level, *service, *nodeID)

	case "stream":
		streamToPino(ctx, *redisAddr)

	case "pino":
		// 直接管道到pino-pretty命令
		pipeToPino(ctx, *redisAddr)

	case "rules-analyze":
		// 规则分析（第一道分析）
		rulesAnalyze(ctx, *redisAddr, *count, *watch, *persist)

	case "llm-analyze":
		// 大模型分析（第二道分析）
		llmAnalyze(ctx, *redisAddr, *ollamaURL, *model, *count, *persist)

	case "rules":
		showRules()

	case "stats":
		showStats(ctx, *redisAddr)

	case "history":
		// 查询历史分析结果
		showHistory(*hours, *ruleFilter)

	case "rule-stats":
		// 查询规则统计
		showRuleStats(*hours)

	case "test":
		testLog(ctx, *redisAddr)

	case "help":
		printHelp()

	default:
		fmt.Printf("未知操作: %s\n", *action)
		printHelp()
		os.Exit(1)
	}
}

// tailLogs 查看最新日志
func tailLogs(ctx context.Context, redisAddr string, count int, level, service, nodeID string) {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	msgs, err := rdb.XRevRangeN(ctx, "logs:stream", "+", "-", int64(count)).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取日志失败: %v\n", err)
		return
	}

	// 将字符串级别转换为Pino数字级别
	levelNum := levelStringToNum(level)

	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		if data, ok := msg.Values["data"].(string); ok {
			var entry map[string]interface{}
			json.Unmarshal([]byte(data), &entry)

			// 过滤级别（支持数字和字符串）
			if level != "" && !matchLevel(entry["level"], levelNum) {
				continue
			}
			// 过滤服务
			svc := entry["svc"]
			if svc == nil {
				svc = entry["name"]
			}
			if service != "" && fmt.Sprintf("%v", svc) != service {
				continue
			}
			// 过滤节点
			node := entry["node"]
			if node == nil {
				node = entry["hostname"]
			}
			if nodeID != "" && fmt.Sprintf("%v", node) != nodeID {
				continue
			}

			printLogEntry(entry)
		}
	}
}

// levelStringToNum 将字符串级别转换为Pino数字
func levelStringToNum(level string) int {
	switch strings.ToLower(level) {
	case "trace":
		return 10
	case "debug":
		return 20
	case "info":
		return 30
	case "warn", "warning":
		return 40
	case "error":
		return 50
	case "fatal":
		return 60
	default:
		return -1
	}
}

// matchLevel 匹配日志级别
func matchLevel(logLevel interface{}, targetLevel int) bool {
	switch v := logLevel.(type) {
	case float64:
		return int(v) == targetLevel
	case int:
		return v == targetLevel
	case string:
		return levelStringToNum(v) == targetLevel
	default:
		return false
	}
}

// followLogs 实时跟踪日志
func followLogs(ctx context.Context, redisAddr string, level, service, nodeID string) {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	lastID := "$"

	// 将字符串级别转换为Pino数字级别
	levelNum := levelStringToNum(level)

	fmt.Println("实时日志跟踪 (按 Ctrl+C 退出)")
	fmt.Println(strings.Repeat("-", 60))

	for {
		select {
		case <-ctx.Done():
			return
		default:
			result, err := rdb.XRead(ctx, &redis.XReadArgs{
				Streams: []string{"logs:stream", lastID},
				Block:   1000,
				Count:   100,
			}).Result()

			if err != nil && err != redis.Nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			for _, stream := range result {
				for _, msg := range stream.Messages {
					if data, ok := msg.Values["data"].(string); ok {
						var entry map[string]interface{}
						json.Unmarshal([]byte(data), &entry)

						// 过滤级别
						if level != "" && !matchLevel(entry["level"], levelNum) {
							continue
						}
						// 过滤服务
						svc := entry["svc"]
						if svc == nil {
							svc = entry["name"]
						}
						if service != "" && fmt.Sprintf("%v", svc) != service {
							continue
						}
						// 过滤节点
						node := entry["node"]
						if node == nil {
							node = entry["hostname"]
						}
						if nodeID != "" && fmt.Sprintf("%v", node) != nodeID {
							continue
						}

						printLogEntry(entry)
					}
					lastID = msg.ID
				}
			}
		}
	}
}

// streamToPino 将日志流式输出（可与pino-pretty配合使用）
func streamToPino(ctx context.Context, redisAddr string) {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	lastID := "0"

	for {
		select {
		case <-ctx.Done():
			return
		default:
			result, _ := rdb.XRead(ctx, &redis.XReadArgs{
				Streams: []string{"logs:stream", lastID},
				Block:   1000,
				Count:   100,
			}).Result()

			for _, stream := range result {
				for _, msg := range stream.Messages {
					if data, ok := msg.Values["data"].(string); ok {
						// 直接输出JSON格式，可以被pino-pretty解析
						fmt.Println(data)
					}
					lastID = msg.ID
				}
			}
		}
	}
}

// pipeToPino 直接管道到pino-pretty命令
func pipeToPino(ctx context.Context, redisAddr string) {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	// 使用npx运行pino-pretty（支持本地和全局安装）
	cmd := exec.CommandContext(ctx, "npx", "pino-pretty", "-t")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建管道失败: %v\n", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "启动pino-pretty失败: %v\n", err)
		fmt.Fprintln(os.Stderr, "请确保已安装: npm install -g pino pino-pretty")
		return
	}

	lastID := "0"
	for {
		select {
		case <-ctx.Done():
			stdin.Close()
			cmd.Wait()
			return
		default:
			result, _ := rdb.XRead(ctx, &redis.XReadArgs{
				Streams: []string{"logs:stream", lastID},
				Block:   1000,
				Count:   100,
			}).Result()

			for _, stream := range result {
				for _, msg := range stream.Messages {
					if data, ok := msg.Values["data"].(string); ok {
						stdin.Write([]byte(data + "\n"))
					}
					lastID = msg.ID
				}
			}
		}
	}
}

// rulesAnalyze 规则分析（第一道分析）
func rulesAnalyze(ctx context.Context, redisAddr string, count int, watch, persist bool) {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	analyzer := rules.NewAnalyzer()

	if watch {
		// 持续监控分析
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		if persist {
			fmt.Println("规则分析 - 持续监控模式 (持久化到PostgreSQL)")
		} else {
			fmt.Println("规则分析 - 持续监控模式 (按 Ctrl+C 退出)")
		}
		fmt.Println(strings.Repeat("=", 60))

		// 立即执行第一次分析
		performRulesAnalysis(ctx, rdb, analyzer, int64(count), persist)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				performRulesAnalysis(ctx, rdb, analyzer, int64(count), persist)
			}
		}
	} else {
		// 单次分析
		performRulesAnalysis(ctx, rdb, analyzer, int64(count), persist)
	}
}

// performRulesAnalysis 执行规则分析
func performRulesAnalysis(ctx context.Context, rdb *redis.Client, analyzer *rules.Analyzer, count int64, persist bool) {
	msgs, err := rdb.XRevRangeN(ctx, "logs:stream", "+", "-", count).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取日志失败: %v\n", err)
		return
	}

	var logs []map[string]interface{}
	for _, msg := range msgs {
		if data, ok := msg.Values["data"].(string); ok {
			var entry map[string]interface{}
			json.Unmarshal([]byte(data), &entry)
			logs = append(logs, entry)
		}
	}

	if len(logs) == 0 {
		fmt.Println("没有日志可分析")
		return
	}

	// 规则分析
	results := analyzer.AnalyzeBatch(logs)
	stats := analyzer.GetStats(results)

	fmt.Printf("\n=== 规则分析报告 ===\n")
	fmt.Printf("分析时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("分析日志数: %d, 匹配规则数: %d\n", len(logs), len(results))
	fmt.Println(strings.Repeat("-", 50))

	// 显示统计
	severityDist := stats["by_severity"].(map[string]int)
	typeDist := stats["by_type"].(map[string]int)
	fmt.Printf("严重级别分布: %v\n", severityDist)
	fmt.Printf("类型分布: %v\n", typeDist)

	// 显示匹配的规则
	var ruleMatches []storage.RuleMatch
	if len(results) > 0 {
		fmt.Println("\n匹配的规则:")
		for i, result := range results {
			severityIcon := getSeverityIcon(result.Severity)
			fmt.Printf("  %s [%s] %s\n", severityIcon, result.RuleID, result.RuleName)
			fmt.Printf("     建议: %s\n", result.Suggestion)
			if result.NeedsLLM {
				fmt.Printf("     ⚠ 需要大模型进一步分析: %s\n", result.Reason)
			}

			// 准备持久化数据
			if persist && i < len(logs) {
				log := logs[i]
				rm := storage.RuleMatch{
					RuleID:     result.RuleID,
					RuleName:   result.RuleName,
					Severity:   string(result.Severity),
					RuleType:   string(result.Type),
					Suggestion: result.Suggestion,
					LogMessage: fmt.Sprintf("%v", log["msg"]),
					LogService: fmt.Sprintf("%v", log["svc"]),
					LogNode:    fmt.Sprintf("%v", log["node"]),
					LogTime:    time.Now(),
					LogFields:  log,
				}
				ruleMatches = append(ruleMatches, rm)
			}
		}
	} else {
		fmt.Println("\n没有匹配任何规则")
	}

	// 持久化到PostgreSQL
	if persist && pgStorage != nil {
		resultID, err := pgStorage.SaveAnalysisResult(len(logs), len(results), severityDist, typeDist, ruleMatches)
		if err != nil {
			fmt.Fprintf(os.Stderr, "持久化分析结果失败: %v\n", err)
		} else {
			fmt.Printf("\n✓ 分析结果已持久化 (ID: %d)\n", resultID)
		}
	}

	// 提示需要LLM分析的日志
	if stats["needs_llm"].(int) > 0 {
		fmt.Printf("\n提示: 有 %d 条日志需要大模型深度分析\n", stats["needs_llm"])
		fmt.Println("使用 'logctl -action llm-analyze' 进行深度分析")
	}
}

// llmAnalyze 大模型分析（第二道分析）
func llmAnalyze(ctx context.Context, redisAddr, ollamaURL, model string, count int, persist bool) {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	analyzer := rules.NewAnalyzer()

	msgs, err := rdb.XRevRangeN(ctx, "logs:stream", "+", "-", int64(count)).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取日志失败: %v\n", err)
		return
	}

	var logs []map[string]interface{}
	for _, msg := range msgs {
		if data, ok := msg.Values["data"].(string); ok {
			var entry map[string]interface{}
			json.Unmarshal([]byte(data), &entry)
			logs = append(logs, entry)
		}
	}

	if len(logs) == 0 {
		fmt.Println("没有日志可分析")
		return
	}

	// 先进行规则分析找出需要LLM分析的日志
	results := analyzer.AnalyzeBatch(logs)
	stats := analyzer.GetStats(results)

	fmt.Printf("\n=== 大模型深度分析 ===\n")
	fmt.Printf("分析时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("分析日志数: %d\n", len(logs))
	fmt.Println(strings.Repeat("-", 50))

	// 收集错误日志和需要LLM分析的日志
	var logsForLLM []map[string]interface{}
	for i, result := range results {
		if result.NeedsLLM {
			logsForLLM = append(logsForLLM, logs[i])
		}
	}

	// 如果没有标记需要LLM的日志，则分析所有错误日志
	if len(logsForLLM) == 0 {
		for i, log := range logs {
			level := pinoLevelToString(log["level"])
			if level == "error" {
				logsForLLM = append(logsForLLM, logs[i])
			}
		}
	}

	if len(logsForLLM) == 0 {
		fmt.Println("没有需要深度分析的错误日志")
		return
	}

	fmt.Printf("需要深度分析的日志: %d 条\n", len(logsForLLM))
	fmt.Printf("规则分析统计: %v\n\n", stats["by_severity"])

	// 调用大模型
	llmContent := callOllama(ctx, ollamaURL, model, logsForLLM)

	// 持久化大模型分析结果
	if persist && pgStorage != nil && llmContent != "" {
		// 先保存规则分析结果
		severityDist := stats["by_severity"].(map[string]int)
		typeDist := stats["by_type"].(map[string]int)
		resultID, err := pgStorage.SaveAnalysisResult(len(logs), len(results), severityDist, typeDist, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "持久化分析结果失败: %v\n", err)
		} else {
			// 保存大模型分析结果
			err = pgStorage.SaveLLMAnalysis(resultID, model, llmContent, len(logsForLLM))
			if err != nil {
				fmt.Fprintf(os.Stderr, "持久化大模型分析结果失败: %v\n", err)
			} else {
				fmt.Printf("\n✓ 大模型分析结果已持久化 (规则结果ID: %d)\n", resultID)
			}
		}
	}
}

// analyzeLogs 分析日志（已废弃，保留向后兼容）
func analyzeLogs(ctx context.Context, redisAddr, ollamaURL, model string, count int, watch bool) {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	analyzer := rules.NewAnalyzer()

	if watch {
		// 持续监控分析
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		fmt.Println("持续监控分析模式 (按 Ctrl+C 退出)")
		fmt.Println(strings.Repeat("=", 60))

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				performAnalysis(ctx, rdb, analyzer, ollamaURL, model, int64(count))
			}
		}
	} else {
		// 单次分析
		performAnalysis(ctx, rdb, analyzer, ollamaURL, model, int64(count))
	}
}

// performAnalysis 执行分析（已废弃，保留向后兼容）
func performAnalysis(ctx context.Context, rdb *redis.Client, analyzer *rules.Analyzer, ollamaURL, model string, count int64) {
	msgs, err := rdb.XRevRangeN(ctx, "logs:stream", "+", "-", count).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取日志失败: %v\n", err)
		return
	}

	var logs []map[string]interface{}
	for _, msg := range msgs {
		if data, ok := msg.Values["data"].(string); ok {
			var entry map[string]interface{}
			json.Unmarshal([]byte(data), &entry)
			logs = append(logs, entry)
		}
	}

	if len(logs) == 0 {
		fmt.Println("没有日志可分析")
		return
	}

	// 规则分析
	results := analyzer.AnalyzeBatch(logs)
	stats := analyzer.GetStats(results)

	fmt.Printf("\n分析时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("分析日志数: %d, 匹配规则数: %d\n", len(logs), len(results))
	fmt.Println(strings.Repeat("-", 40))

	// 显示统计
	fmt.Printf("严重级别分布: %v\n", stats["by_severity"])
	fmt.Printf("类型分布: %v\n", stats["by_type"])

	// 显示匹配的规则
	if len(results) > 0 {
		fmt.Println("\n匹配的规则:")
		for _, result := range results {
			severityIcon := getSeverityIcon(result.Severity)
			fmt.Printf("  %s [%s] %s\n", severityIcon, result.RuleID, result.RuleName)
			fmt.Printf("     建议: %s\n", result.Suggestion)
			if result.NeedsLLM {
				fmt.Printf("     ⚠ 需要大模型进一步分析: %s\n", result.Reason)
			}
		}
	}

	// 需要LLM分析的日志
	if stats["needs_llm"].(int) > 0 {
		fmt.Printf("\n需要大模型深度分析的日志: %d 条\n", stats["needs_llm"])
		var logsNeedLLM []map[string]interface{}
		for i, result := range results {
			if result.NeedsLLM {
				logsNeedLLM = append(logsNeedLLM, logs[i])
			}
		}
		if len(logsNeedLLM) > 0 {
			callOllama(ctx, ollamaURL, model, logsNeedLLM)
		}
	}
}

// callOllama 调用Ollama进行分析
func callOllama(ctx context.Context, ollamaURL, model string, logs []map[string]interface{}) string {
	fmt.Println("\n正在调用大模型进行深度分析...")

	// 构建请求
	var logTexts []string
	for i, log := range logs {
		// 简化日志内容，只保留关键信息
		simplified := fmt.Sprintf("服务: %v, 级别: %v, 消息: %v", log["svc"], log["level"], log["msg"])
		if fields, ok := log["fields"].(map[string]interface{}); ok {
			if errMsg, ok := fields["error_msg"]; ok {
				simplified += fmt.Sprintf(", 错误: %v", errMsg)
			}
			if errType, ok := fields["error_type"]; ok {
				simplified += fmt.Sprintf(", 错误类型: %v", errType)
			}
		}
		logTexts = append(logTexts, fmt.Sprintf("日志%d: %s", i+1, simplified))
	}

	prompt := fmt.Sprintf(`你是边缘设备运维专家。请用中文分析以下工业控制系统日志并提供诊断建议。

【日志内容】
%s

【输出要求】
请严格使用中文输出，不要使用英文。按以下格式输出：

## 问题诊断
（用中文描述发现的问题）

## 根本原因
（用中文分析问题的根本原因）

## 解决建议
（用中文提供具体的解决措施）`, strings.Join(logTexts, "\n"))

	reqBody := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.3,  // 降低温度使输出更确定
			"num_predict": 2048, // 增加token数量
		},
	}
	reqData, _ := json.Marshal(reqBody)

	// 发送请求 - 增加超时时间
	client := &http.Client{Timeout: 180 * time.Second}
	url := fmt.Sprintf("%s/api/generate", ollamaURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建请求失败: %v\n", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "调用Ollama失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// 尝试解析响应
	var ollamaResp struct {
		Response   string `json:"response"`
		Thinking   string `json:"thinking"`
		Error      string `json:"error"`
		Message    string `json:"message"`
		DoneReason string `json:"done_reason"`
	}

	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		fmt.Fprintf(os.Stderr, "解析响应失败: %v\n", err)
		return ""
	}

	if ollamaResp.Error != "" {
		fmt.Fprintf(os.Stderr, "Ollama错误: %s\n", ollamaResp.Error)
		return ""
	}

	if ollamaResp.Message != "" {
		fmt.Fprintf(os.Stderr, "Ollama消息: %s\n", ollamaResp.Message)
		return ""
	}

	fmt.Println("\n=== 大模型分析结果 ===")

	// 优先使用response字段，如果没有则使用thinking字段
	var result string
	if ollamaResp.Response != "" {
		fmt.Println(ollamaResp.Response)
		result = ollamaResp.Response
	} else if ollamaResp.Thinking != "" {
		// 从thinking中提取有用内容
		fmt.Println("(从思考过程中提取分析结果)")
		// 尝试提取中文部分（支持新旧格式）
		var idx int
		if idx = strings.Index(ollamaResp.Thinking, "## 问题诊断"); idx == -1 {
			idx = strings.Index(ollamaResp.Thinking, "问题诊断")
		}
		if idx != -1 {
			fmt.Println(ollamaResp.Thinking[idx:])
			result = ollamaResp.Thinking[idx:]
		} else {
			fmt.Println(ollamaResp.Thinking)
			result = ollamaResp.Thinking
		}
	} else {
		fmt.Fprintf(os.Stderr, "未获取到有效响应内容\n")
	}
	return result
}

// showHistory 显示历史分析结果
func showHistory(hours int, ruleFilter string) {
	if pgStorage == nil {
		fmt.Fprintln(os.Stderr, "PostgreSQL未连接")
		return
	}

	fmt.Println("=== 历史分析结果 ===")
	fmt.Printf("查询范围: 最近 %d 小时\n", hours)
	fmt.Println(strings.Repeat("-", 60))

	// 查询最近的分析结果
	results, err := pgStorage.GetRecentResults(100)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询失败: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Println("没有历史分析结果")
		return
	}

	// 过滤并显示结果
	count := 0
	for _, r := range results {
		// 如果指定了规则过滤，检查是否匹配
		if ruleFilter != "" {
			matches, err := pgStorage.GetRuleMatchesByResultID(r.ID)
			if err != nil {
				continue
			}
			hasRule := false
			for _, m := range matches {
				if m.RuleID == ruleFilter {
					hasRule = true
					break
				}
			}
			if !hasRule {
				continue
			}
		}

		fmt.Printf("\n[%s] 分析ID: %d\n", r.AnalysisTime.Format("2006-01-02 15:04:05"), r.ID)
		fmt.Printf("  日志数: %d, 匹配规则数: %d\n", r.LogCount, r.MatchCount)
		fmt.Printf("  严重级别: %v\n", r.SeverityDistribution)
		fmt.Printf("  类型分布: %v\n", r.TypeDistribution)

		// 显示匹配的规则
		matches, _ := pgStorage.GetRuleMatchesByResultID(r.ID)
		if len(matches) > 0 {
			fmt.Println("  匹配规则:")
			for _, m := range matches {
				severityIcon := getSeverityIconByString(m.Severity)
				fmt.Printf("    %s [%s] %s - %s\n", severityIcon, m.RuleID, m.RuleName, m.LogMessage[:min(50, len(m.LogMessage))])
			}
		}
		count++
	}

	fmt.Printf("\n共 %d 条记录\n", count)
}

// showRuleStats 显示规则统计
func showRuleStats(hours int) {
	if pgStorage == nil {
		fmt.Fprintln(os.Stderr, "PostgreSQL未连接")
		return
	}

	fmt.Println("=== 规则匹配统计 ===")
	fmt.Printf("统计范围: 最近 %d 小时\n", hours)
	fmt.Println(strings.Repeat("-", 60))

	stats, err := pgStorage.GetRuleStats(hours)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询失败: %v\n", err)
		return
	}

	if len(stats) == 0 {
		fmt.Println("没有匹配记录")
		return
	}

	// 获取规则信息
	analyzer := rules.NewAnalyzer()
	allRules := analyzer.GetRules()
	ruleMap := make(map[string]string)
	for _, r := range allRules {
		ruleMap[r.ID] = r.Name
	}

	// 显示统计
	fmt.Println("\n规则ID                      匹配次数  规则名称")
	fmt.Println(strings.Repeat("-", 60))
	for ruleID, count := range stats {
		name := ruleMap[ruleID]
		if name == "" {
			name = "未知规则"
		}
		fmt.Printf("%-28s %6d    %s\n", ruleID, count, name)
	}

	// 显示小时统计
	fmt.Println("\n=== 小时统计趋势 ===")
	hourlyStats, err := pgStorage.GetHourlyStats(hours)
	if err == nil && len(hourlyStats) > 0 {
		for _, s := range hourlyStats {
			fmt.Printf("%s - 日志: %d, 错误: %d, 警告: %d\n",
				s.StatHour.Format("2006-01-02 15:00"),
				s.LogCount, s.ErrorCount, s.WarnCount)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getSeverityIconByString 根据字符串获取严重性图标
func getSeverityIconByString(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}

// showRules 显示所有规则
func showRules() {
	analyzer := rules.NewAnalyzer()
	rules := analyzer.GetRules()

	fmt.Println("=== 日志分析规则列表 ===")
	fmt.Println()

	for _, rule := range rules {
		status := "启用"
		if !rule.Enabled {
			status = "禁用"
		}
		fmt.Printf("规则ID: %s\n", rule.ID)
		fmt.Printf("名称: %s\n", rule.Name)
		fmt.Printf("描述: %s\n", rule.Description)
		fmt.Printf("类型: %s | 严重性: %s | 状态: %s\n", rule.Type, rule.Severity, status)
		fmt.Printf("条件: %v\n", rule.Conditions)
		fmt.Println(strings.Repeat("-", 40))
	}
}

// showStats 显示统计信息
func showStats(ctx context.Context, redisAddr string) {
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	// 获取流信息
	info, err := rdb.XInfoStream(ctx, "logs:stream").Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取流信息失败: %v\n", err)
		return
	}

	fmt.Println("=== Redis日志流统计 ===")
	fmt.Printf("流长度: %d 条\n", info.Length)
	fmt.Printf("第一个消息ID: %s\n", info.FirstEntry.ID)
	fmt.Printf("最后一个消息ID: %s\n", info.LastEntry.ID)

	// 获取最近的日志分析
	msgs, _ := rdb.XRevRangeN(ctx, "logs:stream", "+", "-", 100).Result()

	levelCount := make(map[string]int)
	serviceCount := make(map[string]int)
	nodeCount := make(map[string]int)

	for _, msg := range msgs {
		if data, ok := msg.Values["data"].(string); ok {
			var entry map[string]interface{}
			json.Unmarshal([]byte(data), &entry)

			// 处理级别（支持数字和字符串）
			if level := entry["level"]; level != nil {
				levelStr := pinoLevelToString(level)
				levelCount[levelStr]++
			}
			// 处理服务名
			if svc := entry["svc"]; svc != nil {
				serviceCount[fmt.Sprintf("%v", svc)]++
			} else if name := entry["name"]; name != nil {
				serviceCount[fmt.Sprintf("%v", name)]++
			}
			// 处理节点
			if node := entry["node"]; node != nil {
				nodeCount[fmt.Sprintf("%v", node)]++
			} else if hostname := entry["hostname"]; hostname != nil {
				nodeCount[fmt.Sprintf("%v", hostname)]++
			}
		}
	}

	fmt.Println("\n最近100条日志统计:")
	fmt.Printf("级别分布: %v\n", levelCount)
	fmt.Printf("服务分布: %v\n", serviceCount)
	fmt.Printf("节点分布: %v\n", nodeCount)
}

// testLog 发送测试日志
func testLog(ctx context.Context, redisAddr string) {
	log, err := logger.NewSimpleLogger(redisAddr, "test-node", "test-service")
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建日志器失败: %v\n", err)
		return
	}
	defer log.Close()

	log.Info(ctx, "测试信息日志", map[string]interface{}{
		"test_key": "test_value",
		"number":   123,
	})

	log.Warn(ctx, "测试警告日志", map[string]interface{}{
		"warning": "这是一个测试警告",
	})

	log.Error(ctx, "测试错误日志", map[string]interface{}{
		"error": "这是一个测试错误",
		"retry": 3,
	})

	fmt.Println("测试日志已发送到Redis")
}

// printLogEntry 打印日志条目
func printLogEntry(entry map[string]interface{}) {
	// 解析时间（兼容新旧格式）
	var timestamp time.Time
	if t, ok := entry["time"].(string); ok {
		// 新格式：Pino兼容的ISO时间
		timestamp, _ = time.Parse(time.RFC3339, t)
	} else if ts, ok := entry["ts"].(float64); ok {
		// 旧格式：毫秒时间戳
		timestamp = time.UnixMilli(int64(ts))
	} else {
		timestamp = time.Now()
	}

	// 解析级别（兼容数字和字符串）
	var levelStr string
	if level, ok := entry["level"].(float64); ok {
		// 新格式：Pino数字级别
		levelStr = pinoLevelToString(int(level))
	} else if level, ok := entry["level"].(string); ok {
		levelStr = level
	}

	svc := entry["svc"]
	if svc == nil {
		svc = entry["name"]
	}
	node := entry["node"]
	if node == nil {
		node = entry["hostname"]
	}
	msg := entry["msg"]

	levelIcon := getLevelIcon(levelStr)
	fmt.Printf("[%s] %s | %s | %s | %s",
		timestamp.Format("15:04:05.000"),
		levelIcon,
		svc,
		node,
		msg,
	)

	// 打印字段
	if fields, ok := entry["fields"].(map[string]interface{}); ok && len(fields) > 0 {
		fmt.Printf(" | %v", fields)
	}
	fmt.Println()
}

// pinoLevelToString 将Pino级别转换为字符串（支持数字和字符串输入）
func pinoLevelToString(level interface{}) string {
	switch v := level.(type) {
	case int:
		return pinoNumToString(v)
	case float64:
		return pinoNumToString(int(v))
	case string:
		return v
	default:
		return "info"
	}
}

// pinoNumToString 将Pino数字级别转换为字符串
func pinoNumToString(level int) string {
	switch level {
	case 10:
		return "trace"
	case 20:
		return "debug"
	case 30:
		return "info"
	case 40:
		return "warn"
	case 50:
		return "error"
	case 60:
		return "fatal"
	default:
		return "info"
	}
}

// getLevelIcon 获取日志级别图标
func getLevelIcon(level string) string {
	switch level {
	case "info":
		return "INFO "
	case "warn":
		return "WARN "
	case "error":
		return "ERROR"
	case "debug":
		return "DEBUG"
	case "trace":
		return "TRACE"
	case "fatal":
		return "FATAL"
	default:
		return level
	}
}

// getSeverityIcon 获取严重性图标
func getSeverityIcon(severity rules.Severity) string {
	switch severity {
	case rules.SeverityCritical:
		return "🔴"
	case rules.SeverityHigh:
		return "🟠"
	case rules.SeverityMedium:
		return "🟡"
	case rules.SeverityLow:
		return "🟢"
	default:
		return "⚪"
	}
}

// printHelp 打印帮助信息
func printHelp() {
	fmt.Println(`
边缘设备日志管理工具 - logctl

用法:
  logctl -action <操作> [选项]

操作:
  tail           查看最新日志
  follow         实时跟踪日志
  stream         将日志流式输出（可与pino-pretty配合）
  pino           直接管道到pino-pretty命令格式化输出

  rules-analyze  规则分析（第一道分析）
  llm-analyze    大模型深度分析（第二道分析）

  history        查询历史分析结果
  rule-stats     显示规则匹配统计

  rules          显示所有分析规则
  stats          显示日志统计信息
  test           发送测试日志
  help           显示帮助信息

选项:
  -redis         Redis地址 (默认: localhost:6379)
  -count         显示日志数量 (默认: 20)
  -level         过滤日志级别 (info/warn/error)
  -service       过滤服务名称
  -node          过滤节点ID
  -ollama        Ollama服务地址 (默认: http://localhost:11434)
  -model         Ollama模型名称 (默认: qwen3.5:9b)
  -watch         持续监控模式
  
  持久化选项:
  -persist       持久化分析结果到PostgreSQL
  -pg-host       PostgreSQL主机 (默认: localhost)
  -pg-port       PostgreSQL端口 (默认: 5432)
  -pg-user       PostgreSQL用户 (默认: postgres)
  -pg-password   PostgreSQL密码 (默认: Ba0sight)
  -pg-database   PostgreSQL数据库 (默认: edge_logs)
  
  历史查询选项:
  -hours         查询历史的小时数 (默认: 24)
  -rule          过滤规则ID

示例:
  # 查看最新日志
  logctl -action tail
  logctl -action tail -count 50
  logctl -action tail -level error

  # 实时跟踪日志
  logctl -action follow

  # 规则分析（第一道分析）
  logctl -action rules-analyze
  logctl -action rules-analyze -count 100
  logctl -action rules-analyze -watch    # 持续监控
  logctl -action rules-analyze -watch -persist  # 持续监控并持久化

  # 大模型深度分析（第二道分析）
  logctl -action llm-analyze
  logctl -action llm-analyze -model qwen3.5:9b -count 20
  logctl -action llm-analyze -persist   # 持久化分析结果

  # 查询历史分析结果
  logctl -action history -hours 24
  logctl -action history -rule PLC-001   # 过滤特定规则

  # 查看规则统计
  logctl -action rule-stats -hours 48

  # 使用Pino格式化输出
  logctl -action pino
  logctl -action stream | npx pino-pretty

  # 查看规则和统计
  logctl -action rules
  logctl -action stats
`)
}
