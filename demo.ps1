# 边缘侧分布式日志管理系统演示脚本 (Windows PowerShell)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  边缘侧分布式日志管理系统演示" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查Redis
Write-Host "[1] 检查Redis连接..." -ForegroundColor Yellow
try {
    $redisTest = redis-cli ping 2>$null
    if ($redisTest -eq "PONG") {
        Write-Host "✓ Redis连接正常" -ForegroundColor Green
    }
} catch {
    Write-Host "✗ Redis未运行，请先启动Redis" -ForegroundColor Red
    Write-Host "  Docker命令: docker run -d --name redis -p 6379:6379 redis:7-alpine"
    exit 1
}

# 检查Ollama
Write-Host "[2] 检查Ollama服务..." -ForegroundColor Yellow
try {
    $ollamaTest = Invoke-WebRequest -Uri "http://localhost:11434/api/tags" -UseBasicParsing -TimeoutSec 2
    Write-Host "✓ Ollama服务正常" -ForegroundColor Green
} catch {
    Write-Host "⚠ Ollama未运行，大模型分析功能将不可用" -ForegroundColor Yellow
    Write-Host "  启动命令: ollama serve"
}

# 检查Pino
Write-Host "[3] 检查Pino..." -ForegroundColor Yellow
$pinoAvailable = $false
try {
    $pinoVersion = pino-pretty --version 2>$null
    Write-Host "✓ pino-pretty已安装: $pinoVersion" -ForegroundColor Green
    $pinoAvailable = $true
} catch {
    try {
        $npxVersion = npx --version 2>$null
        Write-Host "⚠ pino-pretty未找到，但npx可用" -ForegroundColor Yellow
        Write-Host "  可使用: npx pino-pretty" -ForegroundColor Yellow
    } catch {
        Write-Host "⚠ pino-pretty未安装" -ForegroundColor Yellow
        Write-Host "  安装命令: npm install -g pino pino-pretty" -ForegroundColor Yellow
    }
}

# 创建bin目录
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin" | Out-Null
}

# 构建程序
Write-Host ""
Write-Host "[4] 编译Go程序..." -ForegroundColor Yellow
go build -o bin/generator-1.exe ./cmd/generator-1
go build -o bin/generator-2.exe ./cmd/generator-2
go build -o bin/logctl.exe ./cmd/logctl
go build -o bin/analyzer.exe ./cmd/analyzer

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ 编译成功" -ForegroundColor Green
} else {
    Write-Host "✗ 编译失败" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  演示场景" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "1. 启动两个日志产生器（snap7drv PLC驱动 + 驾驶任务计算）"
Write-Host "2. 使用logctl查看和分析日志"
Write-Host "3. 使用pino-pretty格式化输出日志"
Write-Host "4. 规则分析检测异常"
Write-Host "5. 大模型深度分析复杂问题"
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 清理旧进程
Get-Process -Name "generator-1","generator-2" -ErrorAction SilentlyContinue | Stop-Process -Force

# 启动日志产生器
Write-Host "启动日志产生器..." -ForegroundColor Yellow
Write-Host "  - snap7drv PLC驱动 (edge-snap7)"
Write-Host "  - 驾驶任务计算 (edge-driving)"
Write-Host ""

$job1 = Start-Job -ScriptBlock { Set-Location $using:PWD; .\bin\generator-1.exe }
$job2 = Start-Job -ScriptBlock { Set-Location $using:PWD; .\bin\generator-2.exe }

Start-Sleep -Seconds 3

Write-Host ""
Write-Host "日志产生器已启动" -ForegroundColor Green
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  演示命令" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host " # 查看最新日志"
Write-Host " .\bin\logctl.exe -action tail"
Write-Host ""
Write-Host " # 实时跟踪日志"
Write-Host " .\bin\logctl.exe -action follow"
Write-Host ""
Write-Host " # 只查看错误日志"
Write-Host " .\bin\logctl.exe -action tail -level error"
Write-Host ""
Write-Host " # 使用Pino格式化输出"
Write-Host " .\bin\logctl.exe -action pino"
Write-Host ""
Write-Host " # 规则分析（第一道分析）"
Write-Host " .\bin\logctl.exe -action rules-analyze"
Write-Host ""
Write-Host " # 大模型深度分析（第二道分析）"
Write-Host " .\bin\logctl.exe -action llm-analyze"
Write-Host ""
Write-Host " # 查看日志统计"
Write-Host " .\bin\logctl.exe -action stats"
Write-Host ""
Write-Host " # 查看分析规则"
Write-Host " .\bin\logctl.exe -action rules"
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Read-Host "按Enter键开始演示"

# 演示1: 查看日志
Write-Host ""
Write-Host "[演示1] 查看最新10条日志" -ForegroundColor Yellow
Write-Host "命令: .\bin\logctl.exe -action tail -count 10"
Write-Host ""
.\bin\logctl.exe -action tail -count 10
Write-Host ""
Read-Host "按Enter继续"

# 演示2: Pino格式化
Write-Host ""
Write-Host "[演示2] Pino格式化输出" -ForegroundColor Yellow
Write-Host "命令: .\bin\logctl.exe -action pino"
Write-Host ""
$result = & .\bin\logctl.exe -action pino 2>&1 | Select-Object -First 20
$result | ForEach-Object { Write-Host $_ }
Write-Host ""

# 演示3: 规则分析
Write-Host ""
Write-Host "[演示3] 规则分析" -ForegroundColor Yellow
Write-Host "命令: .\bin\logctl.exe -action rules-analyze -count 30"
Write-Host ""
.\bin\logctl.exe -action rules-analyze -count 30
Write-Host ""
Read-Host "按Enter继续"

# 演示4: 大模型分析
Write-Host ""
Write-Host "[演示4] 大模型深度分析" -ForegroundColor Yellow
try {
    $ollamaCheck = Invoke-WebRequest -Uri "http://localhost:11434/api/tags" -UseBasicParsing -TimeoutSec 2
    Write-Host "命令: .\bin\logctl.exe -action llm-analyze -count 20"
    Write-Host ""
    .\bin\logctl.exe -action llm-analyze -count 20
} catch {
    Write-Host "Ollama未运行，跳过大模型分析演示" -ForegroundColor Yellow
}
Write-Host ""

# 清理
Write-Host ""
Write-Host "清理..." -ForegroundColor Yellow
Stop-Job $job1, $job2 -ErrorAction SilentlyContinue
Remove-Job $job1, $job2 -ErrorAction SilentlyContinue
Write-Host "演示结束" -ForegroundColor Green
