// CalculateDrivingTask_1 日志产生器 - 模拟驾驶任务计算服务日志
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"edge-log-demo/pkg/logger"
)

// SmartTpc 设备配置
var smartTpcs = []string{"01", "02", "03", "04", "05", "06", "07", "10", "11", "12", "13", "14", "15", "16", "17", "18", "21"}

// Tag 名称配置
var tagNames = []string{"J015", "J069", "R400"}

// 区域前缀
var areaPrefixes = []string{"TPCA", "TPCB", "TPCC"}

func main() {
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	nodeID := getEnv("NODE_ID", "edge-driving")
	serviceName := "CalculateDrivingTask_1"

	log, err := logger.NewSimpleLogger(redisAddr, nodeID, serviceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志器失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	fmt.Printf("[%s] 驾驶任务计算服务日志产生器启动 (节点: %s)\n", serviceName, nodeID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n收到停止信号，正在退出...")
		cancel()
	}()

	rand.NewSource(time.Now().UnixNano() + 1000)
	threadID := 0x15a8

	for {
		select {
		case <-ctx.Done():
			return
		default:
			threadID += rand.Intn(0x100)
			generateDrivingTaskLog(ctx, log, threadID, rand.Intn(100))
			time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)
		}
	}
}

func generateDrivingTaskLog(ctx context.Context, log *logger.SimpleLogger, threadID int, randomVal int) {
	threadHex := fmt.Sprintf("%x", threadID)
	tpcID := smartTpcs[rand.Intn(len(smartTpcs))]
	areaPrefix := areaPrefixes[rand.Intn(len(areaPrefixes))]
	tagName := tagNames[rand.Intn(len(tagNames))]
	fullTagName := fmt.Sprintf("SRL.%s%s.%s", areaPrefix, tpcID, tagName)

	switch {
	case randomVal < 30:
		// 异步读tag值成功 (30%)
		generateReadTagSuccessLog(ctx, log, threadHex, fullTagName)

	case randomVal < 50:
		// 设置SmartTpc状态 (20%)
		generateSetStatusLog(ctx, log, threadHex, tpcID)

	case randomVal < 65:
		// 获取自动状态信息 (15%)
		generateGetAutoStatusLog(ctx, log, threadHex, tpcID)

	case randomVal < 75:
		// 获取空车位信息 (10%)
		generateGetParkingLog(ctx, log, threadHex, tpcID)

	case randomVal < 85:
		// 设备初始化检查 (10%)
		generateInitCheckLog(ctx, log, threadHex, tpcID, areaPrefix)

	case randomVal < 93:
		// 读tag值失败 (8%)
		generateReadTagErrorLog(ctx, log, threadHex, fullTagName)

	case randomVal < 97:
		// 设备状态异常 (4%)
		generateDeviceErrorLog(ctx, log, threadHex, tpcID)

	default:
		// 连接PLC失败 (3%)
		generatePLCErrorLog(ctx, log, threadHex, tpcID)
	}
}

func generateReadTagSuccessLog(ctx context.Context, log *logger.SimpleLogger, threadHex, tagName string) {
	value := rand.Float64() * 10000
	msg := fmt.Sprintf("读tag值：%s 成功，值为：%.6f", tagName, value)

	log.Info(ctx, msg, map[string]interface{}{
		"raw_log":   fmt.Sprintf("[INFO][%s] %s", threadHex, msg),
		"operation": "read_tag",
		"tag_name":  tagName,
		"value":     value,
		"success":   true,
		"thread_id": threadHex,
	})
}

func generateSetStatusLog(ctx context.Context, log *logger.SimpleLogger, threadHex, tpcID string) {
	j069 := rand.Intn(2)
	r400 := rand.Intn(2)
	msg := fmt.Sprintf("更新SmartTpc编号为%s的自动状态信息成功", tpcID)

	log.Info(ctx, msg, map[string]interface{}{
		"raw_log":   fmt.Sprintf("[INFO][%s] %s", threadHex, msg),
		"operation": "set_status",
		"tpc_id":    tpcID,
		"J069":      j069,
		"R400":      r400,
		"success":   true,
		"thread_id": threadHex,
	})
}

func generateGetAutoStatusLog(ctx context.Context, log *logger.SimpleLogger, threadHex, tpcID string) {
	status := rand.Intn(5) + 1
	msg := fmt.Sprintf("获取SmartTpc编号为%s的自动状态信息成功", tpcID)

	log.Info(ctx, msg, map[string]interface{}{
		"raw_log":     fmt.Sprintf("[INFO][%s] %s", threadHex, msg),
		"operation":   "get_auto_status",
		"tpc_id":      tpcID,
		"auto_status": status,
		"success":     true,
		"thread_id":   threadHex,
	})
}

func generateGetParkingLog(ctx context.Context, log *logger.SimpleLogger, threadHex, tpcID string) {
	parkingCount := rand.Intn(20)
	msg := fmt.Sprintf("获取SmartTpc编号为%s的空车位信息成功", tpcID)

	log.Info(ctx, msg, map[string]interface{}{
		"raw_log":       fmt.Sprintf("[INFO][%s] %s", threadHex, msg),
		"operation":     "get_parking",
		"tpc_id":        tpcID,
		"parking_count": parkingCount,
		"success":       true,
		"thread_id":     threadHex,
	})
}

func generateInitCheckLog(ctx context.Context, log *logger.SimpleLogger, threadHex, tpcID, areaPrefix string) {
	j069 := 1
	r400 := 1
	msg := fmt.Sprintf("检查设备%s%s状态正常，J069=1，R400=1", areaPrefix, tpcID)

	log.Info(ctx, msg, map[string]interface{}{
		"raw_log":   fmt.Sprintf("[INFO][%s] %s", threadHex, msg),
		"operation": "init_check",
		"tpc_id":    tpcID,
		"area":      areaPrefix,
		"J069":      j069,
		"R400":      r400,
		"status":    "normal",
		"thread_id": threadHex,
	})
}

func generateReadTagErrorLog(ctx context.Context, log *logger.SimpleLogger, threadHex, tagName string) {
	errorCode := rand.Intn(5) + 13060
	msg := fmt.Sprintf("读tag值：%s 失败，错误码：%d", tagName, errorCode)

	log.Error(ctx, msg, map[string]interface{}{
		"raw_log":    fmt.Sprintf("[ERROR][%s] %s", threadHex, msg),
		"operation":  "read_tag",
		"tag_name":   tagName,
		"error_code": errorCode,
		"error_type": "tag_read_failed",
		"success":    false,
		"thread_id":  threadHex,
	})
}

func generateDeviceErrorLog(ctx context.Context, log *logger.SimpleLogger, threadHex, tpcID string) {
	errorMsg := "设备处于非远程/自动状态"
	msg := fmt.Sprintf("检查设备%s失败，设备处于非远程/自动状态", tpcID)

	log.Warn(ctx, msg, map[string]interface{}{
		"raw_log":    fmt.Sprintf("[INFO][%s] %s", threadHex, msg),
		"operation":  "device_check",
		"tpc_id":     tpcID,
		"error_type": "device_status_error",
		"error_msg":  errorMsg,
		"success":    false,
		"thread_id":  threadHex,
	})
}

func generatePLCErrorLog(ctx context.Context, log *logger.SimpleLogger, threadHex, tpcID string) {
	msg := fmt.Sprintf("连接PLC设备%s超时，请检查网络连接", tpcID)

	log.Error(ctx, msg, map[string]interface{}{
		"raw_log":    fmt.Sprintf("[ERROR][%s] %s", threadHex, msg),
		"operation":  "plc_connect",
		"tpc_id":     tpcID,
		"error_type": "connection_timeout",
		"error_msg":  "PLC connection timeout",
		"success":    false,
		"thread_id":  threadHex,
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
