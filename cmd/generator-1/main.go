// snap7drv 日志产生器 - 模拟西门子PLC S7驱动日志
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"edge-log-demo/pkg/logger"
)

// 设备配置
var devices = []struct {
	name string
	ip   string
	ip2  string
	rack int
	slot int
}{
	{"Device0", "172.16.11.21", "172.16.11.22", 0, 3},
	{"Device1", "172.16.11.23", "", 0, 2},
}

// 数据块配置
var dbBlocks = []int{232, 238, 218, 135, 105, 106}

func main() {
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	nodeID := getEnv("NODE_ID", "edge-snap7")
	serviceName := "snap7drv"

	log, err := logger.NewSimpleLogger(redisAddr, nodeID, serviceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志器失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	fmt.Printf("[%s] snap7drv PLC驱动日志产生器启动 (节点: %s)\n", serviceName, nodeID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n收到停止信号，正在退出...")
		cancel()
	}()

	rand.NewSource(time.Now().UnixNano())
	threadID := int64(0x7f4800000000)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			threadID += rand.Int63n(0x10000)
			generateSnap7Log(ctx, log, threadID, rand.Intn(100))
			time.Sleep(time.Duration(10+rand.Intn(90)) * time.Millisecond)
		}
	}
}

func generateSnap7Log(ctx context.Context, log *logger.SimpleLogger, threadID int64, randomVal int) {
	device := devices[rand.Intn(len(devices))]
	dbNum := dbBlocks[rand.Intn(len(dbBlocks))]

	switch {
	case randomVal < 60:
		// 正常写操作 (60%)
		generateWriteLog(ctx, log, threadID, device, dbNum)

	case randomVal < 75:
		// 写成功日志 (15%)
		generateWriteSuccessLog(ctx, log, threadID, device, dbNum)

	case randomVal < 85:
		// 正常读操作 (10%)
		generateReadLog(ctx, log, threadID, device, dbNum)

	case randomVal < 92:
		// 连接超时错误 (7%)
		generateConnectErrorLog(ctx, log, threadID, device)

	case randomVal < 97:
		// 地址超出范围错误 (5%)
		generateAddressErrorLog(ctx, log, threadID, device, dbNum)

	default:
		// 读取失败错误 (3%)
		generateReadErrorLog(ctx, log, threadID, device, dbNum)
	}
}

func generateWriteLog(ctx context.Context, log *logger.SimpleLogger, threadID int64, device struct {
	name, ip, ip2 string
	rack, slot    int
}, dbNum int) {
	threadHex := fmt.Sprintf("%x", threadID)
	byteOffset := rand.Intn(100)
	bitOffset := rand.Intn(8)
	writeLen := rand.Intn(3) + 1
	writeData := generateHexData(writeLen)

	msg := fmt.Sprintf("CTRL:Devive[%s] Block[DB%d] StartAddr[0] ByteOffset[%d] BitOffset[%d] Write[%s] WriteLen[%d]",
		device.name, dbNum, byteOffset, bitOffset, writeData, writeLen)

	log.Info(ctx, msg, map[string]interface{}{
		"raw_log":     fmt.Sprintf("[INFO][%s][0][OnWriteCmd:2193] %s", threadHex, msg),
		"device":      device.name,
		"block":       fmt.Sprintf("DB%d", dbNum),
		"operation":   "write",
		"byte_offset": byteOffset,
		"write_len":   writeLen,
		"thread_id":   threadHex,
		"ip":          device.ip,
	})
}

func generateWriteSuccessLog(ctx context.Context, log *logger.SimpleLogger, threadID int64, device struct {
	name, ip, ip2 string
	rack, slot    int
}, dbNum int) {
	threadHex := fmt.Sprintf("%x", threadID)
	costTime := rand.Intn(60) + 5

	var connStr string
	if device.ip2 != "" {
		connStr = fmt.Sprintf("ip=%s;ip2=%s;conntype=pg;multilink=1;", device.ip, device.ip2)
	} else {
		connStr = fmt.Sprintf("ip=%s;conntype=pg;multilink=1;", device.ip)
	}

	msg := fmt.Sprintf("Device[%s:%s] DB[%d] Write Success!cost time %d ms",
		connStr, device.name, dbNum, costTime)

	log.Info(ctx, msg, map[string]interface{}{
		"raw_log":   fmt.Sprintf("[INFO][%s][0][OnWriteCmd:2501] %s", threadHex, msg),
		"device":    device.name,
		"block":     fmt.Sprintf("DB%d", dbNum),
		"operation": "write_success",
		"cost_ms":   costTime,
		"ip":        device.ip,
		"thread_id": threadHex,
	})
}

func generateReadLog(ctx context.Context, log *logger.SimpleLogger, threadID int64, device struct {
	name, ip, ip2 string
	rack, slot    int
}, dbNum int) {
	threadHex := fmt.Sprintf("%x", threadID)
	byteOffset := rand.Intn(100)
	readLen := rand.Intn(10) + 1

	msg := fmt.Sprintf("CTRL:Devive[%s] Block[DB%d] StartAddr[0] ByteOffset[%d] ReadLen[%d]",
		device.name, dbNum, byteOffset, readLen)

	log.Info(ctx, msg, map[string]interface{}{
		"raw_log":     fmt.Sprintf("[INFO][%s][0][OnReadCmd:1800] %s", threadHex, msg),
		"device":      device.name,
		"block":       fmt.Sprintf("DB%d", dbNum),
		"operation":   "read",
		"byte_offset": byteOffset,
		"read_len":    readLen,
		"thread_id":   threadHex,
		"ip":          device.ip,
	})
}

func generateConnectErrorLog(ctx context.Context, log *logger.SimpleLogger, threadID int64, device struct {
	name, ip, ip2 string
	rack, slot    int
}) {
	threadHex := fmt.Sprintf("%x", threadID)

	var connStr string
	if device.ip2 != "" {
		connStr = fmt.Sprintf("ip=%s;ip2=%s;conntype=pg;multilink=1;", device.ip, device.ip2)
	} else {
		connStr = fmt.Sprintf("ip=%s;conntype=pg;multilink=1;", device.ip)
	}

	errorType := rand.Intn(3)
	var errorMsg string
	switch errorType {
	case 0:
		errorMsg = "ISO : An error occurred during recv TCP : Connection timed out"
	case 1:
		errorMsg = "TCP : Connection refused"
	case 2:
		errorMsg = "ISO : An error occurred during send TCP : Broken pipe"
	}

	msg := fmt.Sprintf("Connect Failed!Device:[%s %s] rack:%d slot:%d ,Result:[ %s]",
		connStr, device.name, device.rack, device.slot, errorMsg)

	log.Error(ctx, msg, map[string]interface{}{
		"raw_log":    fmt.Sprintf("[ERROR][%s][655470][S7_ConnectToPLC:2070] %s", threadHex, msg),
		"device":     device.name,
		"operation":  "connect",
		"error_type": "connection_failed",
		"error_msg":  errorMsg,
		"ip":         device.ip,
		"rack":       device.rack,
		"slot":       device.slot,
		"thread_id":  threadHex,
	})
}

func generateAddressErrorLog(ctx context.Context, log *logger.SimpleLogger, threadID int64, device struct {
	name, ip, ip2 string
	rack, slot    int
}, dbNum int) {
	threadHex := fmt.Sprintf("%x", threadID)

	var connStr string
	if device.ip2 != "" {
		connStr = fmt.Sprintf("ip=%s;ip2=%s;conntype=pg;multilink=1;", device.ip, device.ip2)
	} else {
		connStr = fmt.Sprintf("ip=%s;conntype=pg;multilink=1;", device.ip)
	}

	msg := fmt.Sprintf("SOLO:Device[%s:]:[%s]:[DB%d] BlockType[DB] XBNumber[%d] Read Failed!Result:[CPU : Address out of range]",
		connStr, device.name, dbNum, dbNum)

	log.Error(ctx, msg, map[string]interface{}{
		"raw_log":    fmt.Sprintf("[ERROR][%s][9437184][UpDateData:1540] %s", threadHex, msg),
		"device":     device.name,
		"block":      fmt.Sprintf("DB%d", dbNum),
		"operation":  "read",
		"error_type": "address_out_of_range",
		"error_msg":  "CPU : Address out of range",
		"ip":         device.ip,
		"thread_id":  threadHex,
	})
}

func generateReadErrorLog(ctx context.Context, log *logger.SimpleLogger, threadID int64, device struct {
	name, ip, ip2 string
	rack, slot    int
}, dbNum int) {
	threadHex := fmt.Sprintf("%x", threadID)

	var connStr string
	if device.ip2 != "" {
		connStr = fmt.Sprintf("ip=%s;ip2=%s;conntype=pg;multilink=1;", device.ip, device.ip2)
	} else {
		connStr = fmt.Sprintf("ip=%s;conntype=pg;multilink=1;", device.ip)
	}

	errorType := rand.Intn(2)
	var errorMsg string
	switch errorType {
	case 0:
		errorMsg = "CPU : No response from PLC"
	case 1:
		errorMsg = "ISO : An error occurred during recv TCP : Connection reset by peer"
	}

	msg := fmt.Sprintf("SOLO:Device[%s:]:[%s]:[DB%d] Read Failed!Result:[%s]",
		connStr, device.name, dbNum, errorMsg)

	log.Error(ctx, msg, map[string]interface{}{
		"raw_log":    fmt.Sprintf("[ERROR][%s][9437184][UpDateData:1540] %s", threadHex, msg),
		"device":     device.name,
		"block":      fmt.Sprintf("DB%d", dbNum),
		"operation":  "read",
		"error_type": "read_failed",
		"error_msg":  errorMsg,
		"ip":         device.ip,
		"thread_id":  threadHex,
	})
}

func generateHexData(length int) string {
	hexChars := []string{"00", "01", "0a", "0f", "1a", "1f", "21", "30", "31", "ff", "45", "4b", "55", "e1", "e2"}
	result := make([]string, length)
	for i := 0; i < length; i++ {
		result[i] = hexChars[rand.Intn(len(hexChars))]
	}
	return strings.Join(result, " ")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
