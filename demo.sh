#!/bin/bash

# 边缘侧分布式日志管理系统演示脚本

echo "========================================"
echo "  边缘侧分布式日志管理系统演示"
echo "========================================"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 检查Redis
echo -e "${YELLOW}[1] 检查Redis连接...${NC}"
redis-cli ping > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Redis连接正常${NC}"
else
    echo -e "${RED}✗ Redis未运行，请先启动Redis${NC}"
    echo "  Docker命令: docker run -d --name redis -p 6379:6379 redis:7-alpine"
    exit 1
fi

# 检查Ollama
echo -e "${YELLOW}[2] 检查Ollama服务...${NC}"
curl -s http://localhost:11434/api/tags > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Ollama服务正常${NC}"
    # 检查模型
    if curl -s http://localhost:11434/api/tags | grep -q "qwen"; then
        echo -e "${GREEN}✓ 模型已安装${NC}"
    else
        echo -e "${YELLOW}⚠ 模型未安装，请运行: ollama pull qwen2.5:7b${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Ollama未运行，大模型分析功能将不可用${NC}"
    echo "  启动命令: ollama serve"
fi

# 检查Pino
echo -e "${YELLOW}[3] 检查Pino...${NC}"
if command -v pino-pretty &> /dev/null; then
    echo -e "${GREEN}✓ pino-pretty已安装${NC}"
elif command -v npx &> /dev/null; then
    echo -e "${GREEN}✓ npx可用（可使用 npx pino-pretty）${NC}"
else
    echo -e "${YELLOW}⚠ pino-pretty未安装${NC}"
    echo "  安装命令: npm install -g pino pino-pretty"
fi

# 构建程序
echo ""
echo -e "${YELLOW}[4] 编译Go程序...${NC}"
go build -o bin/generator-1 ./cmd/generator-1
go build -o bin/generator-2 ./cmd/generator-2
go build -o bin/logctl ./cmd/logctl
go build -o bin/analyzer ./cmd/analyzer

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ 编译成功${NC}"
else
    echo -e "${RED}✗ 编译失败${NC}"
    exit 1
fi

echo ""
echo "========================================"
echo "  演示场景"
echo "========================================"
echo ""
echo "1. 启动两个日志产生器（snap7drv PLC驱动 + 驾驶任务计算）"
echo "2. 使用logctl查看和分析日志"
echo "3. 使用pino-pretty格式化输出日志"
echo "4. 规则分析检测异常"
echo "5. 大模型深度分析复杂问题"
echo ""
echo "========================================"
echo ""

# 清理旧进程
pkill -f "bin/generator" 2>/dev/null

# 启动日志产生器
echo -e "${YELLOW}启动日志产生器...${NC}"
echo "  - snap7drv PLC驱动 (edge-snap7)"
echo "  - 驾驶任务计算 (edge-driving)"
echo ""

./bin/generator-1 &
PID1=$!
./bin/generator-2 &
PID2=$!

sleep 2

echo ""
echo -e "${GREEN}日志产生器已启动 (PID: $PID1, $PID2)${NC}"
echo ""
echo "========================================"
echo "  演示命令"
echo "========================================"
echo ""
echo " # 查看最新日志"
echo " ./bin/logctl -action tail"
echo ""
echo " # 实时跟踪日志"
echo " ./bin/logctl -action follow"
echo ""
echo " # 只查看错误日志"
echo " ./bin/logctl -action tail -level error"
echo ""
echo " # 使用Pino格式化输出"
echo " ./bin/logctl -action pino"
echo ""
echo " # 规则分析（第一道分析）"
echo " ./bin/logctl -action rules-analyze"
echo ""
echo " # 大模型深度分析（第二道分析）"
echo " ./bin/logctl -action llm-analyze"
echo ""
echo " # 查看日志统计"
echo " ./bin/logctl -action stats"
echo ""
echo " # 查看分析规则"
echo " ./bin/logctl -action rules"
echo ""
echo "========================================"
echo ""
echo -e "${YELLOW}按Enter键开始演示...${NC}"
read

# 演示1: 查看日志
echo ""
echo -e "${YELLOW}[演示1] 查看最新10条日志${NC}"
echo "命令: ./bin/logctl -action tail -count 10"
echo ""
./bin/logctl -action tail -count 10
echo ""
echo -e "${YELLOW}按Enter继续...${NC}"
read

# 演示2: Pino格式化
echo ""
echo -e "${YELLOW}[演示2] Pino格式化输出${NC}"
echo "命令: ./bin/logctl -action pino"
echo ""
./bin/logctl -action pino | head -20
echo ""

# 演示3: 规则分析
echo ""
echo -e "${YELLOW}[演示3] 规则分析${NC}"
echo "命令: ./bin/logctl -action rules-analyze -count 30"
echo ""
./bin/logctl -action rules-analyze -count 30
echo ""
echo -e "${YELLOW}按Enter继续...${NC}"
read

# 演示4: 大模型分析（如果Ollama可用）
echo ""
echo -e "${YELLOW}[演示4] 大模型深度分析${NC}"
if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "命令: ./bin/logctl -action llm-analyze -count 20"
    echo ""
    ./bin/logctl -action llm-analyze -count 20
else
    echo -e "${YELLOW}Ollama未运行，跳过大模型分析演示${NC}"
fi
echo ""

# 清理
echo ""
echo -e "${YELLOW}清理...${NC}"
kill $PID1 $PID2 2>/dev/null
echo -e "${GREEN}演示结束${NC}"
