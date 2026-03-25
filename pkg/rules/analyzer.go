// Package rules 提供基于规则的日志分析功能（第一道分析）
package rules

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// RuleType 规则类型
type RuleType string

const (
	RuleTypeError   RuleType = "error"
	RuleTypeWarning RuleType = "warning"
	RuleTypeInfo    RuleType = "info"
	RuleTypeAnomaly RuleType = "anomaly"
)

// Severity 严重级别
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
)

// Rule 分析规则
type Rule struct {
	ID          string
	Name        string
	Description string
	Type        RuleType
	Severity    Severity
	Conditions  []Condition
	Actions     []Action
	Enabled     bool
}

// Condition 规则条件
type Condition struct {
	Field    string      // 字段名: level, msg, svc, node, fields.xxx
	Operator string      // 操作符: eq, neq, contains, matches, gt, lt, gte, lte
	Value    interface{} // 比较值
}

// Action 规则动作
type Action struct {
	Type       string            // alert, escalate, tag
	Parameters map[string]string // 动作参数
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	RuleID     string                 `json:"rule_id"`
	RuleName   string                 `json:"rule_name"`
	Type       RuleType               `json:"type"`
	Severity   Severity               `json:"severity"`
	Message    string                 `json:"message"`
	Suggestion string                 `json:"suggestion"`
	Tags       []string               `json:"tags"`
	NeedsLLM   bool                   `json:"needs_llm"` // 是否需要大模型进一步分析
	Reason     string                 `json:"reason"`    // 需要LLM的原因
	Metadata   map[string]interface{} `json:"metadata"`
}

// Analyzer 规则分析器
type Analyzer struct {
	rules []Rule
}

// NewAnalyzer 创建规则分析器
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		rules: getDefaultRules(),
	}
}

// getDefaultRules 获取默认规则集
func getDefaultRules() []Rule {
	return []Rule{
		// ========== snap7drv PLC驱动规则 ==========
		// PLC连接失败规则
		{
			ID:          "PLC-001",
			Name:        "PLC连接失败",
			Description: "检测西门子PLC连接超时或失败",
			Type:        RuleTypeError,
			Severity:    SeverityCritical,
			Conditions: []Condition{
				{Field: "level", Operator: "eq", Value: "error"},
				{Field: "msg", Operator: "contains", Value: "Connect Failed"},
			},
			Actions: []Action{
				{Type: "alert", Parameters: map[string]string{"channel": "plc-team"}},
				{Type: "escalate", Parameters: map[string]string{"priority": "p1"}},
			},
			Enabled: true,
		},
		// PLC地址错误规则
		{
			ID:          "PLC-002",
			Name:        "PLC地址超出范围",
			Description: "检测PLC数据块地址超出范围错误",
			Type:        RuleTypeError,
			Severity:    SeverityHigh,
			Conditions: []Condition{
				{Field: "level", Operator: "eq", Value: "error"},
				{Field: "msg", Operator: "contains", Value: "Address out of range"},
			},
			Actions: []Action{
				{Type: "alert", Parameters: map[string]string{"channel": "plc-team"}},
			},
			Enabled: true,
		},
		// PLC读取失败规则
		{
			ID:          "PLC-003",
			Name:        "PLC读取失败",
			Description: "检测PLC数据读取失败",
			Type:        RuleTypeError,
			Severity:    SeverityHigh,
			Conditions: []Condition{
				{Field: "level", Operator: "eq", Value: "error"},
				{Field: "fields.error_type", Operator: "eq", Value: "read_failed"},
			},
			Actions: []Action{
				{Type: "alert", Parameters: map[string]string{"channel": "plc-team"}},
			},
			Enabled: true,
		},
		// PLC写入延迟警告
		{
			ID:          "PLC-004",
			Name:        "PLC写入延迟",
			Description: "检测PLC写入操作延迟过高",
			Type:        RuleTypeWarning,
			Severity:    SeverityMedium,
			Conditions: []Condition{
				{Field: "level", Operator: "eq", Value: "info"},
				{Field: "fields.cost_ms", Operator: "gte", Value: 50},
			},
			Actions: []Action{
				{Type: "alert", Parameters: map[string]string{"channel": "plc-team"}},
			},
			Enabled: true,
		},

		// ========== CalculateDrivingTask_1 规则 ==========
		// Tag读取失败规则
		{
			ID:          "TAG-001",
			Name:        "Tag读取失败",
			Description: "检测SmartTpc Tag值读取失败",
			Type:        RuleTypeError,
			Severity:    SeverityHigh,
			Conditions: []Condition{
				{Field: "level", Operator: "eq", Value: "error"},
				{Field: "msg", Operator: "contains", Value: "读tag值"},
				{Field: "msg", Operator: "contains", Value: "失败"},
			},
			Actions: []Action{
				{Type: "alert", Parameters: map[string]string{"channel": "driving-team"}},
			},
			Enabled: true,
		},
		// 设备状态异常规则
		{
			ID:          "TAG-002",
			Name:        "设备状态异常",
			Description: "检测SmartTpc设备处于非远程/自动状态",
			Type:        RuleTypeWarning,
			Severity:    SeverityMedium,
			Conditions: []Condition{
				{Field: "level", Operator: "eq", Value: "warn"},
				{Field: "fields.error_type", Operator: "eq", Value: "device_status_error"},
			},
			Actions: []Action{
				{Type: "alert", Parameters: map[string]string{"channel": "driving-team"}},
			},
			Enabled: true,
		},
		// PLC连接超时规则
		{
			ID:          "TAG-003",
			Name:        "PLC连接超时",
			Description: "检测SmartTpc PLC连接超时",
			Type:        RuleTypeError,
			Severity:    SeverityCritical,
			Conditions: []Condition{
				{Field: "level", Operator: "eq", Value: "error"},
				{Field: "fields.error_type", Operator: "eq", Value: "connection_timeout"},
			},
			Actions: []Action{
				{Type: "alert", Parameters: map[string]string{"channel": "driving-team"}},
				{Type: "escalate", Parameters: map[string]string{"priority": "p1"}},
			},
			Enabled: true,
		},

		// ========== 通用规则 ==========
		// 未知错误类型规则（需要LLM分析）
		{
			ID:          "UNKNOWN-001",
			Name:        "未知错误类型",
			Description: "检测未知的错误类型，需要大模型进一步分析",
			Type:        RuleTypeError,
			Severity:    SeverityHigh,
			Conditions: []Condition{
				{Field: "level", Operator: "eq", Value: "error"},
			},
			Actions: []Action{
				{Type: "tag", Parameters: map[string]string{"tag": "needs-llm-analysis"}},
			},
			Enabled: true,
		},
		// 复杂错误模式（需要LLM深入分析）
		{
			ID:          "PATTERN-001",
			Name:        "复杂错误模式",
			Description: "需要上下文分析的复杂错误模式",
			Type:        RuleTypeAnomaly,
			Severity:    SeverityMedium,
			Conditions: []Condition{
				{Field: "level", Operator: "eq", Value: "error"},
			},
			Actions: []Action{
				{Type: "tag", Parameters: map[string]string{"tag": "complex-pattern"}},
			},
			Enabled: true,
		},
	}
}

// AnalyzeLog 分析单条日志
func (a *Analyzer) AnalyzeLog(log map[string]interface{}) *AnalysisResult {
	for _, rule := range a.rules {
		if !rule.Enabled {
			continue
		}

		if a.matchConditions(log, rule.Conditions) {
			result := &AnalysisResult{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				Type:       rule.Type,
				Severity:   rule.Severity,
				Message:    fmt.Sprintf("规则 [%s] 匹配: %s", rule.ID, rule.Description),
				Suggestion: a.getSuggestion(rule),
				Tags:       a.extractTags(rule),
				NeedsLLM:   a.needsLLMAnalysis(rule),
				Reason:     a.getLLMReason(rule),
				Metadata: map[string]interface{}{
					"rule_type": string(rule.Type),
					"timestamp": time.Now().UnixMilli(),
					"log_level": log["level"],
					"log_msg":   log["msg"],
					"log_svc":   log["svc"],
					"log_node":  log["node"],
				},
			}
			return result
		}
	}

	return nil
}

// AnalyzeBatch 批量分析日志
func (a *Analyzer) AnalyzeBatch(logs []map[string]interface{}) []*AnalysisResult {
	results := make([]*AnalysisResult, 0)
	for _, log := range logs {
		if result := a.AnalyzeLog(log); result != nil {
			results = append(results, result)
		}
	}
	return results
}

// matchConditions 检查日志是否匹配所有条件
func (a *Analyzer) matchConditions(log map[string]interface{}, conditions []Condition) bool {
	for _, cond := range conditions {
		if !a.matchCondition(log, cond) {
			return false
		}
	}
	return true
}

// matchCondition 检查单个条件
func (a *Analyzer) matchCondition(log map[string]interface{}, cond Condition) bool {
	value := a.getFieldValue(log, cond.Field)
	if value == nil {
		return cond.Operator == "eq" && cond.Value == nil
	}

	// 特殊处理level字段的比较（支持Bunyan数字级别和字符串级别）
	if cond.Field == "level" {
		return a.matchLevelCondition(value, cond)
	}

	switch cond.Operator {
	case "eq":
		return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", cond.Value)
	case "neq":
		return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", cond.Value)
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", value), fmt.Sprintf("%v", cond.Value))
	case "matches":
		matched, _ := regexp.MatchString(fmt.Sprintf("%v", cond.Value), fmt.Sprintf("%v", value))
		return matched
	case "gt":
		return a.compareNumbers(value, cond.Value) > 0
	case "lt":
		return a.compareNumbers(value, cond.Value) < 0
	case "gte":
		return a.compareNumbers(value, cond.Value) >= 0
	case "lte":
		return a.compareNumbers(value, cond.Value) <= 0
	}
	return false
}

// matchLevelCondition 匹配日志级别条件（支持数字和字符串格式）
func (a *Analyzer) matchLevelCondition(value interface{}, cond Condition) bool {
	// 将value转换为字符串级别
	valueStr := levelToString(value)
	condStr := fmt.Sprintf("%v", cond.Value)

	switch cond.Operator {
	case "eq":
		return valueStr == condStr
	case "neq":
		return valueStr != condStr
	}
	return false
}

// levelToString 将日志级别转换为字符串（支持Bunyan数字和字符串）
func levelToString(level interface{}) string {
	switch v := level.(type) {
	case float64:
		// Bunyan数字级别
		switch int(v) {
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
	case string:
		return v
	default:
		return fmt.Sprintf("%v", level)
	}
}

// getFieldValue 获取日志字段值
func (a *Analyzer) getFieldValue(log map[string]interface{}, field string) interface{} {
	parts := strings.Split(field, ".")
	if len(parts) == 1 {
		return log[field]
	}

	// 处理嵌套字段，如 fields.retry
	current := log
	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}
	return nil
}

// compareNumbers 比较数字
func (a *Analyzer) compareNumbers(aVal, bVal interface{}) int {
	aFloat := toFloat64(aVal)
	bFloat := toFloat64(bVal)
	if aFloat < bFloat {
		return -1
	} else if aFloat > bFloat {
		return 1
	}
	return 0
}

// toFloat64 转换为float64
func toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	default:
		return 0
	}
}

// getSuggestion 获取建议
func (a *Analyzer) getSuggestion(rule Rule) string {
	suggestions := map[string]string{
		// snap7drv PLC驱动规则
		"PLC-001": "检查PLC网络连接、设备电源和通讯参数配置",
		"PLC-002": "检查数据块地址配置是否正确，确认PLC程序中的地址范围",
		"PLC-003": "检查PLC设备状态、通讯电缆和网络稳定性",
		"PLC-004": "优化PLC通讯参数，检查网络延迟和负载情况",
		// CalculateDrivingTask_1 规则
		"TAG-001": "检查Tag配置和PLC通讯状态，确认Tag地址是否正确",
		"TAG-002": "检查SmartTpc设备的远程/自动切换状态，确认操作模式",
		"TAG-003": "检查PLC网络连接和设备电源状态",
		// 通用规则
		"UNKNOWN-001": "需要进行详细分析以确定根本原因",
		"PATTERN-001": "建议检查服务健康状态和网络连接",
	}
	return suggestions[rule.ID]
}

// extractTags 提取标签
func (a *Analyzer) extractTags(rule Rule) []string {
	tags := []string{string(rule.Type)}
	for _, action := range rule.Actions {
		if action.Type == "tag" {
			if tag, ok := action.Parameters["tag"]; ok {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

// needsLLMAnalysis 判断是否需要LLM分析
func (a *Analyzer) needsLLMAnalysis(rule Rule) bool {
	return rule.ID == "UNKNOWN-001" || rule.ID == "PATTERN-001"
}

// getLLMReason 获取需要LLM分析的原因
func (a *Analyzer) getLLMReason(rule Rule) string {
	reasons := map[string]string{
		"UNKNOWN-001": "未知错误类型，规则分析无法确定根本原因，需要大模型进行语义分析",
		"PATTERN-001": "复杂错误模式，需要结合上下文进行深度分析",
	}
	return reasons[rule.ID]
}

// AddRule 添加自定义规则
func (a *Analyzer) AddRule(rule Rule) {
	a.rules = append(a.rules, rule)
}

// GetRules 获取所有规则
func (a *Analyzer) GetRules() []Rule {
	return a.rules
}

// GetStats 获取统计信息
func (a *Analyzer) GetStats(results []*AnalysisResult) map[string]interface{} {
	stats := map[string]interface{}{
		"total_matches": len(results),
		"by_severity":   make(map[string]int),
		"by_type":       make(map[string]int),
		"needs_llm":     0,
		"rules_matched": make(map[string]int),
	}

	severityMap := stats["by_severity"].(map[string]int)
	typeMap := stats["by_type"].(map[string]int)
	ruleMap := stats["rules_matched"].(map[string]int)

	for _, result := range results {
		severityMap[string(result.Severity)]++
		typeMap[string(result.Type)]++
		ruleMap[result.RuleID]++
		if result.NeedsLLM {
			stats["needs_llm"] = stats["needs_llm"].(int) + 1
		}
	}

	return stats
}
