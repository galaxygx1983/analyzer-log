// Package storage 提供 PostgreSQL 持久化存储功能
package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgresConfig PostgreSQL配置
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// PostgresStorage PostgreSQL存储
type PostgresStorage struct {
	db *sql.DB
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	ID                   int64
	AnalysisTime         time.Time
	LogCount             int
	MatchCount           int
	SeverityDistribution map[string]int
	TypeDistribution     map[string]int
	RulesMatched         []string
	CreatedAt            time.Time
}

// RuleMatch 规则匹配详情
type RuleMatch struct {
	ID         int64
	ResultID   int64
	RuleID     string
	RuleName   string
	Severity   string
	RuleType   string
	Suggestion string
	LogMessage string
	LogService string
	LogNode    string
	LogTime    time.Time
	LogFields  map[string]interface{}
	CreatedAt  time.Time
}

// LLMAnalysis 大模型分析结果
type LLMAnalysis struct {
	ID              int64
	ResultID        int64
	ModelName       string
	AnalysisContent string
	LogCount        int
	AnalysisTime    time.Time
	CreatedAt       time.Time
}

// HourlyStats 每小时统计
type HourlyStats struct {
	ID           int64
	StatHour     time.Time
	LogCount     int
	ErrorCount   int
	WarnCount    int
	InfoCount    int
	ServiceStats map[string]int
	NodeStats    map[string]int
	CreatedAt    time.Time
}

// NewPostgresStorage 创建PostgreSQL存储
func NewPostgresStorage(cfg PostgresConfig) (*PostgresStorage, error) {
	// 首先连接到默认的postgres数据库，检查并创建目标数据库
	defaultDSN := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password)

	defaultDB, err := sql.Open("postgres", defaultDSN)
	if err != nil {
		return nil, fmt.Errorf("连接PostgreSQL失败: %w", err)
	}
	defer defaultDB.Close()

	// 检查数据库是否存在
	var exists bool
	err = defaultDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", cfg.Database).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("检查数据库失败: %w", err)
	}

	// 如果数据库不存在，创建它
	if !exists {
		_, err = defaultDB.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Database))
		if err != nil {
			return nil, fmt.Errorf("创建数据库失败: %w", err)
		}
		fmt.Printf("数据库 '%s' 已创建\n", cfg.Database)
	}

	// 连接到目标数据库
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return &PostgresStorage{db: db}, nil
}

// Close 关闭数据库连接
func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}

// InitTables 初始化表结构
func (ps *PostgresStorage) InitTables() error {
	_, err := ps.db.Exec(`
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

		CREATE TABLE IF NOT EXISTS llm_analysis (
			id SERIAL PRIMARY KEY,
			result_id INTEGER REFERENCES analysis_results(id) ON DELETE CASCADE,
			model_name VARCHAR(100) NOT NULL,
			analysis_content TEXT NOT NULL,
			log_count INTEGER NOT NULL,
			analysis_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

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
	`)
	return err
}

// SaveAnalysisResult 保存分析结果
func (ps *PostgresStorage) SaveAnalysisResult(
	logCount, matchCount int,
	severityDist, typeDist map[string]int,
	ruleMatches []RuleMatch,
) (int64, error) {
	var resultID int64

	// 转换为JSON
	severityJSON, _ := json.Marshal(severityDist)
	typeJSON, _ := json.Marshal(typeDist)

	// 提取规则ID列表
	var ruleIDs []string
	for _, rm := range ruleMatches {
		ruleIDs = append(ruleIDs, rm.RuleID)
	}
	rulesJSON, _ := json.Marshal(ruleIDs)

	// 插入主记录
	err := ps.db.QueryRow(`
		INSERT INTO analysis_results (log_count, match_count, severity_distribution, type_distribution, rules_matched)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, logCount, matchCount, severityJSON, typeJSON, rulesJSON).Scan(&resultID)

	if err != nil {
		return 0, fmt.Errorf("插入分析结果失败: %w", err)
	}

	// 插入规则匹配详情
	for _, match := range ruleMatches {
		fieldsJSON, _ := json.Marshal(match.LogFields)
		_, err := ps.db.Exec(`
			INSERT INTO rule_matches (result_id, rule_id, rule_name, severity, rule_type, suggestion, log_message, log_service, log_node, log_time, log_fields)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, resultID, match.RuleID, match.RuleName, match.Severity, match.RuleType, match.Suggestion, match.LogMessage, match.LogService, match.LogNode, match.LogTime, fieldsJSON)
		if err != nil {
			return 0, fmt.Errorf("插入规则匹配详情失败: %w", err)
		}
	}

	return resultID, nil
}

// SaveLLMAnalysis 保存大模型分析结果
func (ps *PostgresStorage) SaveLLMAnalysis(resultID int64, modelName, content string, logCount int) error {
	_, err := ps.db.Exec(`
		INSERT INTO llm_analysis (result_id, model_name, analysis_content, log_count)
		VALUES ($1, $2, $3, $4)
	`, resultID, modelName, content, logCount)
	return err
}

// SaveHourlyStats 保存小时统计
func (ps *PostgresStorage) SaveHourlyStats(
	statHour time.Time,
	logCount, errorCount, warnCount, infoCount int,
	serviceStats, nodeStats map[string]int,
) error {
	serviceJSON, _ := json.Marshal(serviceStats)
	nodeJSON, _ := json.Marshal(nodeStats)

	_, err := ps.db.Exec(`
		INSERT INTO hourly_stats (stat_hour, log_count, error_count, warn_count, info_count, service_stats, node_stats)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (stat_hour) DO UPDATE SET
			log_count = EXCLUDED.log_count,
			error_count = EXCLUDED.error_count,
			warn_count = EXCLUDED.warn_count,
			info_count = EXCLUDED.info_count,
			service_stats = EXCLUDED.service_stats,
			node_stats = EXCLUDED.node_stats
	`, statHour, logCount, errorCount, warnCount, infoCount, serviceJSON, nodeJSON)
	return err
}

// GetRecentResults 获取最近的分析结果
func (ps *PostgresStorage) GetRecentResults(limit int) ([]AnalysisResult, error) {
	rows, err := ps.db.Query(`
		SELECT id, analysis_time, log_count, match_count, severity_distribution, type_distribution, created_at
		FROM analysis_results
		ORDER BY analysis_time DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AnalysisResult
	for rows.Next() {
		var r AnalysisResult
		var severityJSON, typeJSON []byte
		err := rows.Scan(&r.ID, &r.AnalysisTime, &r.LogCount, &r.MatchCount, &severityJSON, &typeJSON, &r.CreatedAt)
		if err != nil {
			continue
		}
		json.Unmarshal(severityJSON, &r.SeverityDistribution)
		json.Unmarshal(typeJSON, &r.TypeDistribution)
		results = append(results, r)
	}
	return results, nil
}

// GetRuleMatchesByResultID 获取指定分析结果的规则匹配详情
func (ps *PostgresStorage) GetRuleMatchesByResultID(resultID int64) ([]RuleMatch, error) {
	rows, err := ps.db.Query(`
		SELECT id, result_id, rule_id, rule_name, severity, rule_type, suggestion, log_message, log_service, log_node, log_time, log_fields, created_at
		FROM rule_matches
		WHERE result_id = $1
		ORDER BY created_at
	`, resultID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []RuleMatch
	for rows.Next() {
		var m RuleMatch
		var fieldsJSON []byte
		var logTime sql.NullTime
		err := rows.Scan(&m.ID, &m.ResultID, &m.RuleID, &m.RuleName, &m.Severity, &m.RuleType, &m.Suggestion, &m.LogMessage, &m.LogService, &m.LogNode, &logTime, &fieldsJSON, &m.CreatedAt)
		if err != nil {
			continue
		}
		if logTime.Valid {
			m.LogTime = logTime.Time
		}
		json.Unmarshal(fieldsJSON, &m.LogFields)
		matches = append(matches, m)
	}
	return matches, nil
}

// GetRuleMatchHistory 获取规则匹配历史
func (ps *PostgresStorage) GetRuleMatchHistory(ruleID string, hours int) ([]RuleMatch, error) {
	rows, err := ps.db.Query(`
		SELECT id, result_id, rule_id, rule_name, severity, rule_type, suggestion, log_message, log_service, log_node, log_time, log_fields, created_at
		FROM rule_matches
		WHERE rule_id = $1 AND created_at >= NOW() - INTERVAL '1 hour' * $2
		ORDER BY created_at DESC
	`, ruleID, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []RuleMatch
	for rows.Next() {
		var m RuleMatch
		var fieldsJSON []byte
		var logTime sql.NullTime
		err := rows.Scan(&m.ID, &m.ResultID, &m.RuleID, &m.RuleName, &m.Severity, &m.RuleType, &m.Suggestion, &m.LogMessage, &m.LogService, &m.LogNode, &logTime, &fieldsJSON, &m.CreatedAt)
		if err != nil {
			continue
		}
		if logTime.Valid {
			m.LogTime = logTime.Time
		}
		json.Unmarshal(fieldsJSON, &m.LogFields)
		matches = append(matches, m)
	}
	return matches, nil
}

// GetHourlyStats 获取小时统计
func (ps *PostgresStorage) GetHourlyStats(hours int) ([]HourlyStats, error) {
	rows, err := ps.db.Query(`
		SELECT id, stat_hour, log_count, error_count, warn_count, info_count, service_stats, node_stats, created_at
		FROM hourly_stats
		WHERE stat_hour >= NOW() - INTERVAL '1 hour' * $1
		ORDER BY stat_hour DESC
	`, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []HourlyStats
	for rows.Next() {
		var s HourlyStats
		var serviceJSON, nodeJSON []byte
		err := rows.Scan(&s.ID, &s.StatHour, &s.LogCount, &s.ErrorCount, &s.WarnCount, &s.InfoCount, &serviceJSON, &nodeJSON, &s.CreatedAt)
		if err != nil {
			continue
		}
		json.Unmarshal(serviceJSON, &s.ServiceStats)
		json.Unmarshal(nodeJSON, &s.NodeStats)
		stats = append(stats, s)
	}
	return stats, nil
}

// GetRuleStats 获取规则统计
func (ps *PostgresStorage) GetRuleStats(hours int) (map[string]int, error) {
	rows, err := ps.db.Query(`
		SELECT rule_id, COUNT(*) as count
		FROM rule_matches
		WHERE created_at >= NOW() - INTERVAL '1 hour' * $1
		GROUP BY rule_id
		ORDER BY count DESC
	`, hours)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var ruleID string
		var count int
		if err := rows.Scan(&ruleID, &count); err != nil {
			continue
		}
		stats[ruleID] = count
	}
	return stats, nil
}
