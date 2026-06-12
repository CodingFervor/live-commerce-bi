<h1 align="center">Live Commerce BI</h1>

<p align="center">
  <strong>Enterprise-grade Business Intelligence Platform for Live-Streaming E-Commerce</strong>
</p>

<p align="center">
  <a href="#english">English</a> | <a href="#中文">中文</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go" alt="Go 1.22" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql" alt="PostgreSQL 16" />
  <img src="https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis" alt="Redis 7" />
  <img src="https://img.shields.io/badge/Kafka-7.6-231F20?style=flat&logo=apachekafka" alt="Kafka 7.6" />
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker" alt="Docker" />
  <img src="https://img.shields.io/badge/License-MIT-green" alt="License" />
</p>

---

<a id="english"></a>

## Overview

Live Commerce BI is an enterprise-grade business intelligence platform designed specifically for live-streaming e-commerce operations. It provides real-time analytics, multi-platform data integration, AI-powered insights, and comprehensive reporting capabilities across major Chinese e-commerce platforms.

### Key Features

- **Multi-Platform Integration** - Unified data from Douyin (TikTok), Kuaishou, Taobao Live, JD Live, and Pinduoduo Live
- **Real-Time Analytics** - WebSocket-powered live dashboards with sub-second metric updates
- **100+ API Endpoints** - Comprehensive RESTful API covering all business domains
- **AI-Powered Intelligence** - Multi-model AI integration (OpenAI, Qwen, DeepSeek, Zhipu GLM, Baidu Wenxin, Moonshot, iFlytek, Ollama)
- **Advanced Analytics** - Cohort analysis, RFM segmentation, sales forecasting, anomaly detection, OLAP queries
- **Fine-Grained RBAC** - 23 permissions with row-level data access control
- **Event Tracking SDK** - Real-time user behavior tracking and funnel analysis
- **Data Quality Engine** - Configurable data quality rules with automated checking
- **Alert System** - Intelligent alerting with multi-channel notifications (Email, SMS, DingTalk, Webhook)
- **Export Engine** - Asynchronous CSV/JSON export with progress tracking

### Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Language | Go | 1.22 |
| Web Framework | Gin | 1.10 |
| Database | PostgreSQL | 16 |
| Cache | Redis | 7 |
| Message Queue | Kafka (segmentio/kafka-go) | 7.6 |
| WebSocket | gorilla/websocket | 1.5 |
| Auth | JWT (golang-jwt/v5) + bcrypt | - |
| Containerization | Docker + Docker Compose | - |

### Architecture

```
                    +-----------+
                    | Frontend  |
                    | (React)   |
                    +-----+-----+
                          |
                    +-----v-----+
                    | API Server |
                    | (Go/Gin)   |
                    +--+--+--+--+
                       |  |  |
            +----------+  |  +----------+
            |             |             |
      +-----v-----+ +-----v-----+ +-----v-----+
      | PostgreSQL| |   Redis   | |   Kafka   |
      |    (16)   | |    (7)    | |  (Events) |
      +-----------+ +-----------+ +-----------+
```

**Project Structure:**

```
live-commerce-bi/
+-- cmd/api/              # Application entry point
+-- internal/
|   +-- config/           # Configuration loader
|   +-- database/         # PostgreSQL connection pool
|   +-- cache/            # Redis client wrapper
|   +-- middleware/        # Auth, CORS, rate limiting, data permissions
|   +-- model/            # Data models (20+ structs)
|   +-- repository/       # Data access layer (parameterized SQL)
|   +-- service/          # Business logic layer
|   +-- handler/          # HTTP handlers (Gin controllers)
+-- pkg/
|   +-- logger/           # Structured logging
|   +-- jwt/              # JWT token management
|   +-- response/         # Unified API response format
|   +-- hash/             # Password hashing utilities
|   +-- pagination/       # Pagination helpers
+-- config/
|   +-- config.json       # Application configuration
+-- sql/
|   +-- init.sql          # Core schema (18 tables, 55 indexes)
|   +-- advanced.sql      # Advanced features (19 tables, RBAC)
|   +-- ai_settings.sql   # AI configuration (3 tables)
+-- docs/
|   +-- API.md            # API documentation
+-- docker-compose.yml    # Full-stack deployment
+-- Dockerfile            # Multi-stage build
+-- Makefile              # Development tools
```

---

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- (Optional) Go 1.22+ for local development

### 1. Clone the Repository

```bash
git clone https://github.com/CodingFervor/live-commerce-bi.git
cd live-commerce-bi
```

### 2. Configure Environment

```bash
# Copy and edit environment variables
cp .env.example .env
```

Edit `.env` with your secrets:

```env
DB_PASSWORD=your_secure_db_password
JWT_SECRET=your_secure_jwt_secret_at_least_32_chars
REDIS_PASSWORD=your_redis_password
```

Edit `config/config.json` to match your environment. For Docker deployment, the defaults work out of the box.

### 3. Launch with Docker Compose

```bash
docker-compose up -d --build
```

This starts all services:
- **App** (port 8080) - The BI API server
- **PostgreSQL** (port 5432) - Database with auto-initialized schema
- **Redis** (port 6379) - Cache layer
- **Kafka** (port 9092) - Event streaming
- **Zookeeper** (port 2181) - Kafka coordination

### 4. Verify the Service

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected"
}
```

### 5. Seed Demo Data (Optional)

```bash
make seed
```

### 6. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

## Local Development

### Without Docker (Manual Setup)

1. Install Go 1.22+, PostgreSQL 16, Redis 7
2. Create database and run schema:

```bash
createdb live_commerce_bi
psql -d live_commerce_bi -f sql/init.sql
psql -d live_commerce_bi -f sql/advanced.sql
psql -d live_commerce_bi -f sql/ai_settings.sql
```

3. Update `config/config.json` with local credentials
4. Run the server:

```bash
# Install dependencies
make deps

# Run in development mode
make dev

# Or build and run
make build
./bin/server
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the binary |
| `make run` | Run the server |
| `make dev` | Run in debug mode |
| `make test` | Run tests with coverage |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code |
| `make deps` | Download dependencies |
| `make docker-up` | Start all Docker services |
| `make docker-down` | Stop all Docker services |
| `make seed` | Insert demo data |
| `make benchmark` | Run benchmarks |

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CONFIG_PATH` | Config file path | `config/config.json` |
| `DB_PASSWORD` | PostgreSQL password | Required |
| `JWT_SECRET` | JWT signing secret | Required |
| `REDIS_PASSWORD` | Redis password | Empty |

### config.json Structure

```json
{
  "server": {
    "port": 8080,
    "mode": "release",
    "read_timeout": 30,
    "write_timeout": 30
  },
  "database": {
    "host": "postgres",
    "port": 5432,
    "user": "biuser",
    "password": "",
    "dbname": "live_commerce_bi",
    "max_open_conns": 50,
    "max_idle_conns": 10,
    "conn_max_lifetime": 3600
  },
  "redis": {
    "host": "redis",
    "port": 6379,
    "password": "",
    "pool_size": 20
  },
  "jwt": {
    "secret": "",
    "expire_hours": 24,
    "issuer": "live-commerce-bi"
  },
  "kafka": {
    "brokers": ["kafka:9092"],
    "group_id": "live-bi-consumer",
    "enabled": false
  },
  "alert": {
    "evaluate_interval": 60,
    "enabled": true
  }
}
```

Database password and JWT secret are resolved from environment variables at runtime.

---

## API Overview

The API is organized into the following route groups:

| Group | Prefix | Endpoints | Description |
|-------|--------|-----------|-------------|
| Auth | `/api/v1/auth` | 3 | Login, register, profile |
| Dashboard | `/api/v1/dashboard` | 3 | Overview, realtime, trends |
| Custom Dashboards | `/api/v1/dashboards` | 8 | CRUD dashboards & widgets |
| Data Sources | `/api/v1/datasources` | 7 | Platform data management |
| Live Rooms | `/api/v1/live-rooms` | 6 | Live stream tracking |
| Streamers | `/api/v1/streamers` | 6 | Streamer management & rankings |
| Products | `/api/v1/products` | 5 | Product management & rankings |
| Orders | `/api/v1/orders` | 4 | Order tracking & statistics |
| Analytics | `/api/v1/analytics` | 6 | GMV, conversion, platform analysis |
| Viewer Analytics | `/api/v1/viewers` | 4 | Demographics, engagement, retention |
| Advanced Analytics | `/api/v1/advanced-analytics` | 10 | Cohort, RFM, forecast, OLAP |
| Reports | `/api/v1/reports` | 9 | Report CRUD & generation |
| Alerts | `/api/v1/alerts` | 7 | Alert rules & notifications |
| Organization | `/api/v1/organizations` | 7 | Multi-tenant org management |
| RBAC | `/api/v1/rbac` | 6 | Role-based access control |
| Audit Logs | `/api/v1/audit-logs` | 2 | Operation audit trail |
| Data Export | `/api/v1/exports` | 5 | Async data export |
| Data Quality | `/api/v1/data-quality` | 4 | Quality rules & checking |
| Event Tracking | `/api/v1/events` | 4 | Behavior tracking SDK |
| Metrics | `/api/v1/metrics` | 3 | Aggregated metrics |
| AI | `/api/v1/ai` | 8 | AI chat, insights, reports |
| System Settings | `/api/v1/system` | 4 | System configuration |
| WebSocket | `/ws` | 3 | Real-time connections |

See [docs/API.md](docs/API.md) for complete endpoint documentation.

---

## AI Integration

Live Commerce BI integrates with 8 AI providers through a unified OpenAI-compatible interface:

| Provider | Default Endpoint | Models |
|----------|-----------------|--------|
| OpenAI | api.openai.com | GPT-4, GPT-3.5 |
| Qwen (Tongyi Qianwen) | dashscope.aliyuncs.com | Qwen-Turbo, Qwen-Plus |
| Zhipu GLM | open.bigmodel.cn | GLM-4, GLM-3-Turbo |
| Baidu Wenxin | aip.baidubce.com | ERNIE-Bot |
| DeepSeek | api.deepseek.com | DeepSeek-Chat |
| Moonshot (Kimi) | api.moonshot.cn | Moonshot-v1 |
| iFlytek Spark | spark-api.xf-yun.com | Spark-V3 |
| Local (Ollama) | localhost:11434 | Any GGUF model |

### AI Features

- **Smart Query** - Natural language to data insights
- **Auto Insights** - AI-generated analysis for any metric
- **Report Generation** - Automated analytical reports
- **Data Explanation** - Plain-language metric interpretation

### Configuring AI

Configure via the AI Settings API:

```bash
# Create an AI provider config
curl -X POST http://localhost:8080/api/v1/ai/configs \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "DeepSeek Chat",
    "provider": "deepseek",
    "api_key": "sk-xxx",
    "model_name": "deepseek-chat",
    "max_tokens": 4096,
    "temperature": 0.7,
    "is_default": true
  }'
```

---

## Security

- **JWT Authentication** with HS256 signing
- **bcrypt** password hashing (cost factor 12)
- **RBAC** with 23 fine-grained permissions
- **Row-Level Data Permissions** - Filter data by platform, streamer, department
- **Rate Limiting** - 200 req/min per IP (Redis-backed)
- **CORS** - Configurable allowed origins
- **SQL Injection Prevention** - Parameterized queries throughout
- **Input Validation** - All endpoints validate request body
- **Audit Logging** - Every data-modifying operation logged
- **API Key Masking** - AI provider keys masked in API responses
- **Path Traversal Protection** - File path validation in exports

---

## Deployment

### Docker (Recommended)

```bash
# Production deployment
docker-compose up -d --build

# View logs
docker-compose logs -f app

# Scale considerations:
# - Increase max_open_conns for high traffic
# - Add Redis replicas for cache HA
# - Use Kafka partitions for event throughput
```

### Environment Variables for Production

```env
DB_PASSWORD=<strong-random-password>
JWT_SECRET=<at-least-32-char-random-string>
REDIS_PASSWORD=<redis-password>
CONFIG_PATH=config/config.json
```

### Database Migrations

Schema is auto-initialized via Docker entrypoint. For updates:

```bash
# Run additional migrations
docker-compose exec postgres psql -U biuser -d live_commerce_bi -f /path/to/migration.sql
```

### Health Monitoring

```bash
# Health check endpoint
curl http://localhost:8080/health

# Docker health status
docker-compose ps
```

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go standard project layout
- Use parameterized SQL queries (no string concatenation)
- Write tests for new functionality
- Run `make lint` before submitting PRs
- Keep API responses consistent with the unified response format

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---
---

<a id="中文"></a>

## 项目简介

Live Commerce BI 是一个专为直播电商打造的企业级商业智能平台，提供实时数据分析、多平台数据集成、AI 智能分析和全面的报表功能，覆盖国内主流电商平台。

### 核心特性

- **多平台数据整合** - 统一接入抖音、快手、淘宝直播、京东直播、拼多多直播数据
- **实时数据看板** - 基于 WebSocket 的实时指标推送，亚秒级更新
- **100+ API 接口** - 覆盖所有业务领域的完整 RESTful API
- **AI 智能分析** - 接入 8 家主流 AI 模型（OpenAI、通义千问、DeepSeek、智谱 GLM、百度文心、Moonshot、讯飞星火、Ollama 本地模型）
- **高级分析引擎** - 留存分析（Cohort）、RFM 用户分层、销售预测、异常检测、OLAP 多维查询
- **精细化权限控制** - 23 项功能权限 + 行级数据权限控制
- **事件追踪 SDK** - 实时用户行为追踪与漏斗分析
- **数据质量引擎** - 可配置的数据质量规则，自动化检查
- **智能告警系统** - 多渠道告警通知（邮件、短信、钉钉、Webhook）
- **数据导出引擎** - 异步 CSV/JSON 导出，支持进度跟踪

### 技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 开发语言 | Go | 1.22 |
| Web 框架 | Gin | 1.10 |
| 数据库 | PostgreSQL | 16 |
| 缓存 | Redis | 7 |
| 消息队列 | Kafka (segmentio/kafka-go) | 7.6 |
| WebSocket | gorilla/websocket | 1.5 |
| 认证 | JWT (golang-jwt/v5) + bcrypt | - |
| 容器化 | Docker + Docker Compose | - |

### 系统架构

```
                    +-----------+
                    |  前端应用  |
                    | (React)   |
                    +-----+-----+
                          |
                    +-----v-----+
                    | API 服务器 |
                    | (Go/Gin)  |
                    +--+--+--+--+
                       |  |  |
            +----------+  |  +----------+
            |             |             |
      +-----v-----+ +-----v-----+ +-----v-----+
      | PostgreSQL| |   Redis   | |   Kafka   |
      |    (16)   | |    (7)    | |  (事件流)  |
      +-----------+ +-----------+ +-----------+
```

**项目结构：**

```
live-commerce-bi/
+-- cmd/api/              # 应用入口
+-- internal/
|   +-- config/           # 配置加载
|   +-- database/         # PostgreSQL 连接池
|   +-- cache/            # Redis 客户端封装
|   +-- middleware/        # 认证、CORS、限流、数据权限
|   +-- model/            # 数据模型（20+ 结构体）
|   +-- repository/       # 数据访问层（参数化 SQL）
|   +-- service/          # 业务逻辑层
|   +-- handler/          # HTTP 处理器（Gin 控制器）
+-- pkg/
|   +-- logger/           # 结构化日志
|   +-- jwt/              # JWT 令牌管理
|   +-- response/         # 统一 API 响应格式
|   +-- hash/             # 密码哈希工具
|   +-- pagination/       # 分页辅助
+-- config/
|   +-- config.json       # 应用配置文件
+-- sql/
|   +-- init.sql          # 核心表结构（18 张表，55 个索引）
|   +-- advanced.sql      # 高级功能（19 张表，RBAC）
|   +-- ai_settings.sql   # AI 配置（3 张表）
+-- docs/
|   +-- API.md            # API 接口文档
+-- docker-compose.yml    # 全栈部署编排
+-- Dockerfile            # 多阶段构建
+-- Makefile              # 开发工具命令
```

---

## 快速开始

### 前提条件

- [Docker](https://docs.docker.com/get-docker/) 和 Docker Compose
- （可选）Go 1.22+ 用于本地开发

### 1. 克隆项目

```bash
git clone https://github.com/CodingFervor/live-commerce-bi.git
cd live-commerce-bi
```

### 2. 配置环境变量

```bash
# 复制环境变量模板
cp .env.example .env
```

编辑 `.env` 文件，填入安全密钥：

```env
DB_PASSWORD=你的数据库密码
JWT_SECRET=你的JWT密钥至少32位
REDIS_PASSWORD=你的Redis密码
```

编辑 `config/config.json` 适配你的环境。Docker 部署使用默认配置即可。

### 3. 使用 Docker Compose 启动

```bash
docker-compose up -d --build
```

启动的服务：
- **App**（端口 8080）- BI API 服务器
- **PostgreSQL**（端口 5432）- 数据库，自动初始化表结构
- **Redis**（端口 6379）- 缓存层
- **Kafka**（端口 9092）- 事件流
- **Zookeeper**（端口 2181）- Kafka 协调服务

### 4. 验证服务

```bash
curl http://localhost:8080/health
```

预期响应：
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected"
}
```

### 5. 导入演示数据（可选）

```bash
make seed
```

### 6. 登录系统

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

---

## 本地开发

### 手动安装（不使用 Docker）

1. 安装 Go 1.22+、PostgreSQL 16、Redis 7
2. 创建数据库并执行建表脚本：

```bash
createdb live_commerce_bi
psql -d live_commerce_bi -f sql/init.sql
psql -d live_commerce_bi -f sql/advanced.sql
psql -d live_commerce_bi -f sql/ai_settings.sql
```

3. 修改 `config/config.json` 中的本地连接信息
4. 启动服务：

```bash
# 安装依赖
make deps

# 开发模式运行
make dev

# 或编译后运行
make build
./bin/server
```

### Makefile 命令

| 命令 | 说明 |
|------|------|
| `make build` | 编译二进制文件 |
| `make run` | 运行服务 |
| `make dev` | 调试模式运行 |
| `make test` | 运行测试并生成覆盖率 |
| `make lint` | 运行代码检查 |
| `make fmt` | 格式化代码 |
| `make deps` | 下载依赖 |
| `make docker-up` | 启动所有 Docker 服务 |
| `make docker-down` | 停止所有 Docker 服务 |
| `make seed` | 插入演示数据 |
| `make benchmark` | 运行性能测试 |

---

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `CONFIG_PATH` | 配置文件路径 | `config/config.json` |
| `DB_PASSWORD` | PostgreSQL 密码 | 必填 |
| `JWT_SECRET` | JWT 签名密钥 | 必填 |
| `REDIS_PASSWORD` | Redis 密码 | 空 |

### config.json 结构

```json
{
  "server": {
    "port": 8080,
    "mode": "release",
    "read_timeout": 30,
    "write_timeout": 30
  },
  "database": {
    "host": "postgres",
    "port": 5432,
    "user": "biuser",
    "password": "",
    "dbname": "live_commerce_bi",
    "max_open_conns": 50,
    "max_idle_conns": 10,
    "conn_max_lifetime": 3600
  },
  "redis": {
    "host": "redis",
    "port": 6379,
    "password": "",
    "pool_size": 20
  },
  "jwt": {
    "secret": "",
    "expire_hours": 24,
    "issuer": "live-commerce-bi"
  },
  "kafka": {
    "brokers": ["kafka:9092"],
    "group_id": "live-bi-consumer",
    "enabled": false
  },
  "alert": {
    "evaluate_interval": 60,
    "enabled": true
  }
}
```

数据库密码和 JWT 密钥在运行时从环境变量读取。

---

## API 概览

API 按以下路由组组织：

| 分组 | 路径前缀 | 接口数 | 说明 |
|------|---------|--------|------|
| 认证 | `/api/v1/auth` | 3 | 登录、注册、个人信息 |
| 数据看板 | `/api/v1/dashboard` | 3 | 总览、实时、趋势 |
| 自定义看板 | `/api/v1/dashboards` | 8 | 看板和组件 CRUD |
| 数据源 | `/api/v1/datasources` | 7 | 平台数据管理 |
| 直播间 | `/api/v1/live-rooms` | 6 | 直播间追踪 |
| 主播 | `/api/v1/streamers` | 6 | 主播管理与排行 |
| 商品 | `/api/v1/products` | 5 | 商品管理与排行 |
| 订单 | `/api/v1/orders` | 4 | 订单追踪与统计 |
| 数据分析 | `/api/v1/analytics` | 6 | GMV、转化率、平台分析 |
| 观众分析 | `/api/v1/viewers` | 4 | 画像、互动、留存 |
| 高级分析 | `/api/v1/advanced-analytics` | 10 | 留存、RFM、预测、OLAP |
| 报表 | `/api/v1/reports` | 9 | 报表增删改查与生成 |
| 告警 | `/api/v1/alerts` | 7 | 告警规则与通知 |
| 组织管理 | `/api/v1/organizations` | 7 | 多租户组织管理 |
| 权限控制 | `/api/v1/rbac` | 6 | 角色权限管理 |
| 审计日志 | `/api/v1/audit-logs` | 2 | 操作审计追踪 |
| 数据导出 | `/api/v1/exports` | 5 | 异步数据导出 |
| 数据质量 | `/api/v1/data-quality` | 4 | 质量规则与检查 |
| 事件追踪 | `/api/v1/events` | 4 | 行为追踪 SDK |
| 指标聚合 | `/api/v1/metrics` | 3 | 聚合指标查询 |
| AI 功能 | `/api/v1/ai` | 8 | AI 对话、洞察、报表 |
| 系统设置 | `/api/v1/system` | 4 | 系统配置管理 |
| WebSocket | `/ws` | 3 | 实时连接 |

完整接口文档见 [docs/API.md](docs/API.md)。

---

## AI 集成

Live Commerce BI 通过统一的 OpenAI 兼容接口接入 8 家 AI 服务商：

| 服务商 | 默认端点 | 模型 |
|--------|---------|------|
| OpenAI | api.openai.com | GPT-4, GPT-3.5 |
| 通义千问 (Qwen) | dashscope.aliyuncs.com | Qwen-Turbo, Qwen-Plus |
| 智谱 GLM | open.bigmodel.cn | GLM-4, GLM-3-Turbo |
| 百度文心 | aip.baidubce.com | ERNIE-Bot |
| DeepSeek | api.deepseek.com | DeepSeek-Chat |
| Moonshot (Kimi) | api.moonshot.cn | Moonshot-v1 |
| 讯飞星火 | spark-api.xf-yun.com | Spark-V3 |
| 本地模型 (Ollama) | localhost:11434 | 任意 GGUF 模型 |

### AI 功能

- **智能查询** - 自然语言转数据洞察
- **自动分析** - AI 自动生成指标分析
- **报表生成** - 自动生成分析报告
- **数据解读** - 通俗语言解读业务指标

### 配置 AI

通过 AI 设置接口配置：

```bash
# 创建 AI 服务商配置
curl -X POST http://localhost:8080/api/v1/ai/configs \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "DeepSeek 对话",
    "provider": "deepseek",
    "api_key": "sk-xxx",
    "model_name": "deepseek-chat",
    "max_tokens": 4096,
    "temperature": 0.7,
    "is_default": true
  }'
```

---

## 安全特性

- **JWT 认证** - HS256 签名令牌
- **bcrypt 加密** - 密码哈希（cost factor 12）
- **RBAC 权限** - 23 项精细功能权限
- **行级数据权限** - 按平台、主播、部门过滤数据
- **频率限制** - 每IP每分钟200次请求（Redis 实现）
- **CORS 控制** - 可配置允许的来源域名
- **SQL 注入防护** - 全局参数化查询
- **输入验证** - 所有接口请求数据验证
- **审计日志** - 所有数据修改操作记录
- **API Key 脱敏** - AI 服务密钥在 API 响应中掩码处理
- **路径遍历防护** - 文件导出路径验证

---

## 部署

### Docker 部署（推荐）

```bash
# 生产环境部署
docker-compose up -d --build

# 查看日志
docker-compose logs -f app

# 扩展建议：
# - 高流量时增加 max_open_conns
# - 增加 Redis 副本实现缓存高可用
# - 使用 Kafka 分区提升事件吞吐
```

### 生产环境变量

```env
DB_PASSWORD=<强随机密码>
JWT_SECRET=<至少32位随机字符串>
REDIS_PASSWORD=<Redis密码>
CONFIG_PATH=config/config.json
```

### 数据库迁移

表结构通过 Docker entrypoint 自动初始化。更新迁移：

```bash
docker-compose exec postgres psql -U biuser -d live_commerce_bi -f /path/to/migration.sql
```

### 健康监控

```bash
# 健康检查接口
curl http://localhost:8080/health

# Docker 容器状态
docker-compose ps
```

---

## 参与贡献

1. Fork 本仓库
2. 创建功能分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 提交 Pull Request

### 开发规范

- 遵循 Go 标准项目布局
- 使用参数化 SQL 查询（禁止字符串拼接）
- 为新功能编写测试
- 提交 PR 前运行 `make lint`
- 保持 API 响应格式统一

---

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。
