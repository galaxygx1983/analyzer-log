// Package analyzer 大模型日志分析器（第二道分析）
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"edge-log-demo/pkg/rules"

	"github.com/redis/go-redis/v9"
)

// LLMAnalyzer 大模型分析器
type LLMAnalyzer struct {
	ollamaURL    string
	model        string
	redis        *redis.Client
	ruleAnalyzer *rules.Analyzer
}

// NewLLMAnalyzer 创建大模型分析器
func NewLLMAnalyzer(ollamaURL, model, redisAddr string) *LLMAnalyzer {
	return &LLMAnalyzer{
		ollamaURL:    ollamaURL,
		model:        model,
		redis:        redis.NewClient(&redis.Options{Addr: redisAddr}),
		ruleAnalyzer: rules.NewAnalyzer(),
	}
}

// LLMResponse 大模型响应结构
type LLMResponse struct {
	HasAnomaly  bool     `json:"has_anomaly"`
	AnomalyType string   `json:"anomaly_type"`
	RootCause   string   `json:"root_cause"`
	Suggestion  string   `json:"suggestion"`
	Confidence  float64  `json:"confidence"`
	RelatedLogs []string `json:"related_logs"`
}

// AnalyzeRecent 分析最近的日志
func (la *LLMAnalyzer) AnalyzeRecent(ctx context.Context, count int64) (string, error) {
	// 从Redis读取日志
	msgs, err := la.redis.XRevRangeN(ctx, "logs:stream", "+", "-", count).Result()
	if err != nil {
		return "", fmt.Errorf("读取Redis失败: %w", err)
	}

	if len(msgs) == 0 {
		return "没有找到日志记录", nil
	}

	// 解析日志
	var logs []map[string]interface{}
	for _, msg := range msgs {
		var entry map[string]interface{}
		if data, ok := msg.Values["data"].(string); ok {
			json.Unmarshal([]byte(data), &entry)
			logs = append(logs, entry)
		}
	}

	// 第一道分析：规则分析
	ruleResults := la.ruleAnalyzer.AnalyzeBatch(logs)
	stats := la.ruleAnalyzer.GetStats(ruleResults)

	// 分离需要LLM分析的日志
	var logsNeedLLM []map[string]interface{}
	var ruleAnalysisResults []map[string]interface{}

	for i, result := range ruleResults {
		if result.NeedsLLM {
			logsNeedLLM = append(logsNeedLLM, logs[i])
		}
		ruleAnalysisResults = append(ruleAnalysisResults, map[string]interface{}{
			"rule_id":    result.RuleID,
			"rule_name":  result.RuleName,
			"severity":   result.Severity,
			"suggestion": result.Suggestion,
		})
	}

	// 构建分析报告
	var report strings.Builder
	report.WriteString("=== 日志分析报告 ===\n\n")
	report.WriteString(fmt.Sprintf("分析时间: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("分析日志数量: %d\n", len(logs)))
	report.WriteString(fmt.Sprintf("规则匹配数量: %d\n\n", len(ruleResults)))

	// 规则分析统计
	report.WriteString("--- 规则分析统计 ---\n")
	report.WriteString(fmt.Sprintf("严重级别分布: %v\n", stats["by_severity"]))
	report.WriteString(fmt.Sprintf("类型分布: %v\n", stats["by_type"]))
	report.WriteString(fmt.Sprintf("规则匹配: %v\n\n", stats["rules_matched"]))

	// 规则分析详情
	if len(ruleAnalysisResults) > 0 {
		report.WriteString("--- 规则分析详情 ---\n")
		for _, r := range ruleAnalysisResults {
			report.WriteString(fmt.Sprintf("[%s] %s (严重性: %s)\n", r["rule_id"], r["rule_name"], r["severity"]))
			report.WriteString(fmt.Sprintf("  建议: %s\n\n", r["suggestion"]))
		}
	}

	// 第二道分析：大模型分析
	if len(logsNeedLLM) > 0 {
		report.WriteString("\n--- 大模型深度分析 ---\n")
		report.WriteString(fmt.Sprintf("需要深度分析的日志数量: %d\n\n", len(logsNeedLLM)))

		llmResult, err := la.callLLM(ctx, logsNeedLLM)
		if err != nil {
			report.WriteString(fmt.Sprintf("大模型分析失败: %v\n", err))
		} else {
			report.WriteString(llmResult)
		}
	} else {
		report.WriteString("\n所有问题已通过规则分析解决，无需大模型深度分析。\n")
	}

	return report.String(), nil
}

// AnalyzeSpecific 分析指定的日志（用于需要LLM的日志）
func (la *LLMAnalyzer) AnalyzeSpecific(ctx context.Context, logEntry map[string]interface{}, ruleResult *rules.AnalysisResult) (string, error) {
	// 构建提示词
	prompt := la.buildPromptForSpecificLog(logEntry, ruleResult)

	// 调用Ollama
	return la.callOllamaAPI(ctx, prompt)
}

// callLLM 调用大模型分析
func (la *LLMAnalyzer) callLLM(ctx context.Context, logs []map[string]interface{}) (string, error) {
	// 构建日志文本
	var logTexts []string
	for i, log := range logs {
		data, _ := json.MarshalIndent(log, "", "  ")
		logTexts = append(logTexts, fmt.Sprintf("日志 %d:\n%s", i+1, string(data)))
	}

	prompt := fmt.Sprintf(`你是一个专业的边缘设备运维助手，请分析以下边缘设备日志，提供详细的诊断和解决方案。

日志内容:
%s

请用中文回答以下内容:
1. 异常概述: 是否有异常？异常类型是什么？
2. 根本原因分析: 可能的根本原因是什么？
3. 影响评估: 这个问题的影响范围和严重程度
4. 解决建议: 具体的解决步骤和建议
5. 预防措施: 如何避免类似问题再次发生

请以结构化的方式回答，每个部分使用清晰的标题。`, strings.Join(logTexts, "\n\n"))

	return la.callOllamaAPI(ctx, prompt)
}

// buildPromptForSpecificLog 构建特定日志的提示词
func (la *LLMAnalyzer) buildPromptForSpecificLog(logEntry map[string]interface{}, ruleResult *rules.AnalysisResult) string {
	data, _ := json.MarshalIndent(logEntry, "", "  ")

	return fmt.Sprintf(`你是一个专业的边缘设备运维助手。规则分析系统已经标记了这条日志需要进一步分析。

日志内容:
%s

规则分析结果:
- 匹配规则: %s (%s)
- 需要LLM分析的原因: %s
- 当前建议: %s

请提供更深入的分析:
1. 这条日志背后可能隐藏的具体问题
2. 建议的排查步骤
3. 紧急处理方案`, string(data), ruleResult.RuleName, ruleResult.RuleID, ruleResult.Reason, ruleResult.Suggestion)
}

// callOllamaAPI 调用Ollama API
func (la *LLMAnalyzer) callOllamaAPI(ctx context.Context, prompt string) (string, error) {
	// 使用HTTP客户端调用Ollama API
	// Ollama API 格式: POST http://localhost:11434/api/generate
	// 或使用OpenAI兼容API: POST http://localhost:11434/v1/chat/completions

	// 使用Ollama原生API
	reqBody := map[string]interface{}{
		"model":  la.model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.7,
			"num_predict": 1024,
		},
	}

	reqData, _ := json.Marshal(reqBody)

	// 这里使用简单的HTTP调用
	// 实际项目中可以使用ollama go客户端
	resp, err := la.ollamaRequest(ctx, reqData)
	if err != nil {
		return "", fmt.Errorf("调用Ollama失败: %w", err)
	}

	return resp, nil
}

// ollamaRequest 发送Ollama请求
func (la *LLMAnalyzer) ollamaRequest(ctx context.Context, reqData []byte) (string, error) {
	// 创建HTTP请求
	client := &http.Client{Timeout: 60 * time.Second}

	url := fmt.Sprintf("%s/api/generate", la.ollamaURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析Ollama响应
	var ollamaResp struct {
		Response string `json:"response"`
		Error    string `json:"error"`
	}

	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if ollamaResp.Error != "" {
		return "", fmt.Errorf("Ollama错误: %s", ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}

// WatchAndAnalyze 持续监控并分析
func (la *LLMAnalyzer) WatchAndAnalyze(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Println("开始持续监控分析...")
	fmt.Printf("分析间隔: %v\n\n", interval)

	for {
		select {
		case <-ticker.C:
			fmt.Printf("\n%s\n", strings.Repeat("=", 50))
			result, err := la.AnalyzeRecent(ctx, 50)
			if err != nil {
				fmt.Printf("分析失败: %v\n", err)
				continue
			}
			fmt.Println(result)
		case <-ctx.Done():
			fmt.Println("\n停止监控分析")
			return
		}
	}
}

// GetRecentLogs 获取最近的日志
func (la *LLMAnalyzer) GetRecentLogs(ctx context.Context, count int64) ([]map[string]interface{}, error) {
	msgs, err := la.redis.XRevRangeN(ctx, "logs:stream", "+", "-", count).Result()
	if err != nil {
		return nil, err
	}

	var logs []map[string]interface{}
	for _, msg := range msgs {
		var entry map[string]interface{}
		if data, ok := msg.Values["data"].(string); ok {
			json.Unmarshal([]byte(data), &entry)
			logs = append(logs, entry)
		}
	}
	return logs, nil
}

func main() {
	// 配置
	ollamaURL := getEnv("OLLAMA_URL", "http://localhost:11434")
	model := getEnv("OLLAMA_MODEL", "qwen2.5:7b")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	mode := getEnv("ANALYZE_MODE", "once") // once 或 watch

	analyzer := NewLLMAnalyzer(ollamaURL, model, redisAddr)
	ctx := context.Background()

	if mode == "watch" {
		// 持续监控模式
		analyzer.WatchAndAnalyze(ctx, 30*time.Second)
	} else {
		// 单次分析模式
		result, err := analyzer.AnalyzeRecent(ctx, 50)
		if err != nil {
			fmt.Fprintf(os.Stderr, "分析失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
