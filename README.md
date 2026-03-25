# Analyzer Log

日志分析工具 - 用于解析和分析工控系统日志文件

## 项目结构

```
analyzer-log/
├── cmd/          # 命令行程序入口
│   ├── analyzer/     # 主分析程序
│   ├── generator-1/  # 日志生成器 1
│   ├── generator-2/  # 日志生成器 2
│   ├── file-collector/ # 文件收集工具
│   └── logctl/       # 日志控制工具
├── pkg/          # 核心包
│   ├── logger/       # 日志模块
│   ├── rules/        # 分析规则
│   └── storage/      # 数据存储（PostgreSQL）
├── web/          # Web 前端
├── skill/        # AI Skills
├── docs/         # 文档
├── bin/          # 编译输出
├── logs/         # 运行日志
└── assets/       # 资源文件
```

## 技术栈

- **后端**: Go 1.25.0
- **前端**: 现代 Web 技术栈
- **数据库**: PostgreSQL
- **缓存**: Redis

## 依赖

```go
require (
    github.com/fsnotify/fsnotify v1.9.0
    github.com/lib/pq v1.12.0
    github.com/redis/go-redis/v9 v9.18.0
)
```

## 快速开始

### 安装依赖

```bash
go mod download
```

### 编译

```bash
go build -o bin/analyzer.exe ./cmd/analyzer
go build -o bin/generator-1.exe ./cmd/generator-1
go build -o bin/generator-2.exe ./cmd/generator-2
go build -o bin/file-collector.exe ./cmd/file-collector
go build -o bin/logctl.exe ./cmd/logctl
```

### 运行

```bash
./bin/analyzer --help
```

## 功能特性

- 日志文件解析与收集
- 规则引擎分析
- 数据存储到 PostgreSQL
- Redis 缓存支持
- Web 界面展示

## 开发

### 运行测试

```bash
go test ./...
```

### 代码格式化

```bash
go fmt ./...
go vet ./...
```

## 许可证

MIT License
