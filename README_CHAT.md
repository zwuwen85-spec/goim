# GoIM Chat - 增强版即时通讯系统

> 基于 [Terry-Mao/goim](https://github.com/Terry-Mao/goim) 开发的 QQ 风格即时通讯系统，支持单聊、群聊、AI 聊天等功能。

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

---

## 功能特性

### 核心功能
- **单聊 (Private Chat)** - 一对一即时消息
- **群聊 (Group Chat)** - 支持多人大群
- **AI 聊天 (AI Chat)** - 集成 OpenAI GPT，支持多种 AI 人格
- **好友系统** - 好友添加、删除、分组管理
- **群组管理** - 创建群组、邀请成员、群设置
- **消息推送** - 基于 WebSocket 的实时推送

### 技术特性
- **高性能** - 支持百万级在线用户
- **可扩展** - 分布式架构，水平扩展
- **多协议** - 支持 WebSocket、TCP
- **消息持久化** - 消息存储在 MySQL
- **上下文管理** - AI 对话上下文记忆

---

## 快速开始

### 环境要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | 1.21+ | 编译项目 |
| MySQL | 8.0+ | 数据存储 |
| Redis | 7.0+ | 缓存 |
| Kafka | 3.0+ | 消息队列 |
| Node.js | 18+ | 前端开发（可选） |

### 使用 Docker 启动依赖服务

```bash
docker-compose up -d
```

### 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE goim_chat CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入表结构
mysql -u root -p goim_chat < scripts/chat-schema.sql

# 导入测试数据（可选）
mysql -u root -p goim_chat < scripts/chat-seed.sql
```

### 编译项目

```bash
# 编译所有组件
go build -o target/comet ./cmd/comet
go build -o target/logic ./cmd/logic
go build -o target/job ./cmd/job
go build -o target/chatapi ./cmd/chatapi
```

### 配置文件

复制配置模板并根据实际环境修改：

```bash
cp cmd/logic/logic-example.toml target/logic.toml
cp cmd/comet/comet-example.toml target/comet.toml
cp cmd/job/job-example.toml target/job.toml

# 编辑 ChatAPI 配置，特别是 MySQL 密码和 OpenAI API Key
vi cmd/chatapi/chatapi.toml
```

**重要配置项：**

```toml
# cmd/chatapi/chatapi.toml

[mysql]
dsn = "root:your_password@tcp(127.0.0.1:3306)/goim_chat?charset=utf8mb4"

[ai]
apiKey = "sk-your-openai-api-key"  # 替换为你的 OpenAI API Key
```

### 启动服务

```bash
# 启动 Logic
./target/logic -conf=target/logic.toml &

# 启动 Comet
./target/comet -conf=target/comet.toml &

# 启动 Job
./target/job -conf=target/job.toml &

# 启动 ChatAPI
./target/chatapi cmd/chatapi/chatapi.toml &
```

### 快速测试

```bash
# 1. 注册用户
curl -X POST http://localhost:3112/api/user/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456","nickname":"测试用户"}'

# 2. 登录获取 Token
TOKEN=$(curl -s -X POST http://localhost:3112/api/user/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}' \
  | jq -r '.data.token')

# 3. 获取 AI Bot 列表
curl -X GET http://localhost:3112/api/ai/bots \
  -H "Authorization: Bearer $TOKEN"

# 4. 发送 AI 消息
curl -X POST http://localhost:3112/api/ai/chat/send \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"bot_id": 9001, "message": "你好，请介绍一下自己"}'
```

---

## 项目结构

```
goim/
├── cmd/                    # 各服务入口
│   ├── comet/             # WebSocket 接入服务
│   ├── logic/             # 业务逻辑服务
│   ├── job/               # 消息推送服务
│   └── chatapi/           # HTTP API 服务 (新增)
├── internal/              # 内部包
│   ├── ai/                # AI 服务模块 (新增)
│   │   ├── bot.go         # Bot 管理器
│   │   ├── context.go     # 对话上下文管理
│   │   ├── openai.go      # OpenAI 集成
│   │   └── service.go     # AI 服务接口
│   └── chatapi/           # ChatAPI 内部实现 (新增)
│       ├── dao/           # 数据访问层
│       ├── handler/       # HTTP 处理器
│       ├── middleware/    # 中间件 (JWT)
│       ├── model/         # 数据模型
│       └── service/       # 业务服务
├── scripts/               # 数据库脚本
│   ├── chat-schema.sql    # 数据库结构
│   └── chat-seed.sql      # 测试数据
├── web/                   # Vue 前端项目
├── test_client.html       # WebSocket 测试页面
└── DEPLOYMENT.md          # 详细部署文档
```

---

## API 接口

### 基础地址
```
http://localhost:3112/api
```

### 认证接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/user/register` | POST | 用户注册 |
| `/user/login` | POST | 用户登录 |
| `/user/profile` | GET | 获取用户信息 |

### 好友接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/friend/request` | POST | 发送好友请求 |
| `/friend/accept/:id` | POST | 接受好友请求 |
| `/friend/reject/:id` | POST | 拒绝好友请求 |
| `/friend/delete/:id` | DELETE | 删除好友 |
| `/friend/list` | GET | 获取好友列表 |

### 群组接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/group/create` | POST | 创建群组 |
| `/group/join/:id` | POST | 加入群组 |
| `/group/leave/:id` | DELETE | 退出群组 |
| `/group/info/:id` | GET | 获取群组信息 |
| `/group/members/:id` | GET | 获取群成员列表 |

### 消息接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/message/send` | POST | 发送消息 |
| `/message/history` | GET | 获取历史消息 |
| `/message/read` | POST | 标记消息已读 |

### AI 聊天接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/ai/bots` | GET | 获取 AI Bot 列表 |
| `/ai/bot/create` | POST | 创建自定义 Bot |
| `/ai/chat/send` | POST | 发送 AI 消息 |

### AI Bot 预设

| Bot ID | 名称 | 描述 |
|--------|------|------|
| 9001 | 智能助手 | 有帮助、知识渊博、礼貌 |
| 9002 | 聊天伙伴 | 友好、有同理心、有趣 |
| 9003 | 学习导师 | 知识渊博、耐心、鼓励学习 |
| 9004 | 创意助手 | 有创造力、启发性、原创 |

---

## 系统架构

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │     │   Client    │     │   Client    │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                    │                    │
       └────────────────────┼────────────────────┘
                            │
                    ┌───────▼────────┐
                    │     Nginx      │
                    └───────┬────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼──────┐  ┌────────▼────────┐  ┌───────▼──────┐
│    Comet     │  │    ChatAPI      │  │    Comet     │
│  (WebSocket) │  │   (HTTP API)    │  │  (WebSocket) │
│   Port 3102  │  │   Port 3112     │  │   Port 3102  │
└───────┬──────┘  └────────┬────────┘  └───────┬──────┘
        │                  │                   │
        └──────────────────┼───────────────────┘
                           │
                  ┌────────▼────────┐
                  │      Logic      │
                  │  Port 3111/3119 │
                  └────────┬────────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                   │
┌───────▼──────┐  ┌────────▼────────┐  ┌───────▼──────┐
│     Job      │  │      Redis       │  │    MySQL     │
│  (Kafka)     │  │   Port 6379      │  │   Port 3306  │
└───────┬──────┘  └──────────────────┘  └──────────────┘
        │
┌───────▼──────┐
│    Kafka     │
│  Port 9092   │
└──────────────┘
```

---

## 配置说明

### 端口说明

| 服务 | 端口 | 协议 | 说明 |
|------|------|------|------|
| Comet | 3101 | TCP | TCP 连接 |
| Comet | 3102 | WebSocket | WebSocket 连接 |
| Comet | 3103 | WSS | WebSocket TLS |
| Logic | 3111 | HTTP | HTTP API |
| Logic | 3119 | gRPC | RPC 服务 |
| ChatAPI | 3112 | HTTP | HTTP API |

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MYSQL_DSN` | MySQL 连接字符串 | - |
| `REDIS_ADDR` | Redis 地址 | `127.0.0.1:6379` |
| `KAFKA_BROKERS` | Kafka 地址 | `127.0.0.1:9092` |
| `OPENAI_API_KEY` | OpenAI API Key | - |

---

## 部署指南

详细的部署指南请参考 [DEPLOYMENT.md](./DEPLOYMENT.md)

### 快速部署

```bash
# 1. 启动依赖服务
docker-compose up -d

# 2. 初始化数据库
mysql -u root -p goim_chat < scripts/chat-schema.sql

# 3. 编译
make build

# 4. 启动服务
make start
```

### 生产部署

```bash
# 使用 systemd 管理
sudo systemctl start goim-comet
sudo systemctl start goim-logic
sudo systemctl start goim-job
sudo systemctl start goim-chatapi
```

---

## 开发指南

### 前端开发

```bash
cd web
npm install
npm run dev
```

### 运行测试

```bash
# 单元测试
go test ./...

# 集成测试
go test -tags=integration ./...
```

### 添加新的 AI Bot

```go
// 在 internal/ai/service.go 中添加
func DefaultPersonalities() map[string]*Personality {
    return map[string]*Personality{
        // ... 现有配置
        "custom": {
            Name: "自定义助手",
            Tone: "friendly",
            Role: "assistant",
            Traits: []string{"smart", "funny"},
            SystemPrompt: "You are a custom assistant...",
        },
    }
}
```

---

## 常见问题

### Q: AI 聊天没有响应？

**A:** 检查以下几点：
1. OpenAI API Key 是否正确配置
2. 网络是否能访问 `api.openai.com`
3. 查看服务日志：`tail -f logs/chatapi.log`

### Q: WebSocket 连接失败？

**A:** 检查：
1. Comet 服务是否启动：`lsof -i :3102`
2. 防火墙是否放行端口
3. 使用测试页面验证：`test_client.html`

### Q: 消息推送延迟？

**A:** 检查：
1. Kafka 是否正常运行
2. Redis 是否正常
3. 网络延迟

---

## 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 许可证

本项目基于 MIT 许可证开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

---

## 致谢

- [Terry-Mao/goim](https://github.com/Terry-Mao/goim) - 核心即时通讯框架
- [OpenAI](https://openai.com/) - AI 服务支持
- [Gin](https://github.com/gin-gonic/gin) - Web 框架

---

## 联系方式

- 项目主页: [https://github.com/your-repo/goim](https://github.com/your-repo/goim)
- 问题反馈: [Issues](https://github.com/your-repo/goim/issues)
