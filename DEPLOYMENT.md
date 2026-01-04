# GoIM Chat 项目启动指南

本文档详细说明如何启动和部署 GoIM Chat 项目。

## 目录

- [项目概述](#项目概述)
- [系统架构](#系统架构)
- [环境要求](#环境要求)
- [安装步骤](#安装步骤)
- [配置说明](#配置说明)
- [启动服务](#启动服务)
- [测试功能](#测试功能)
- [常见问题](#常见问题)

---

## 项目概述

GoIM Chat 是一个基于 GoIM 的高性能即时通讯系统，模仿 QQ 的功能设计，支持：

- **单聊** (Private Chat)
- **群聊** (Group Chat)
- **AI 聊天** (AI Chat with OpenAI)
- **好友系统** (Friend System)
- **群组管理** (Group Management)

### 技术栈

| 组件 | 技术 |
|------|------|
| 后端语言 | Go 1.21+ |
| 消息队列 | Apache Kafka |
| 缓存 | Redis |
| 数据库 | MySQL 8.0 |
| 前端 | Vue 3 + TypeScript |
| WebSocket | GoIM Comet |
| AI 服务 | OpenAI API |

---

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                           客户端层                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │   Web    │  │  Android │  │   iOS    │  │  Desktop │        │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └────┬─────┘        │
└───────┼────────────┼────────────┼────────────┼─────────────────┘
        │            │            │            │
        └────────────┴────────────┴────────────┘
                             │
        ┌────────────────────┴────────────────────┐
        │              负载均衡器 (Nginx)            │
        └────────────────────┬────────────────────┘
                             │
        ┌────────────────────┴────────────────────┐
        │                                           │
┌───────▼────────┐  ┌──────────────────────────────────┐
│   COMET        │  │         ChatAPI (HTTP API)        │
│  (接入层)       │  │                                   │
│  - WebSocket   │  │  - 用户认证                        │
│  - TCP        │  │  - 好友管理                        │
│  - 端口: 3101  │  │  - 群组管理                        │
│  - 端口: 3102  │  │  - 消息发送                        │
└───────┬────────┘  │  - AI 聊天                         │
        │           │  - 端口: 3112                     │
        │           └──────────────┬───────────────────┘
        │                          │
        │           ┌──────────────▼───────────────────┐
        │           │         LOGIC                    │
        │           │  (业务逻辑层)                      │
        │           │  - 消息路由                        │
        │           │  - 端口: 3111 (HTTP)             │
        │           │  - 端口: 3119 (gRPC)             │
        │           └──────────────┬───────────────────┘
        │                          │
        │           ┌──────────────▼───────────────────┐
        │           │         JOB                      │
        │           │  (消息推送层)                      │
        │           │  - Kafka 消费                     │
        │           │  - 批量推送                        │
        │           └──────────────┬───────────────────┘
        │                          │
        └──────────────────────────┼──────────────────┐
                                   │                  │
        ┌──────────────────────────┼──────────────────┼──────────┐
        │                          │                  │          │
┌───────▼──────┐  ┌──────────────────▼─┐  ┌─────────▼─────┐  ┌──▼────────┐
│    Kafka     │  │      Redis         │  │     MySQL     │  │  OpenAI   │
│  (消息队列)    │  │     (缓存)         │  │    (数据库)     │  │  (AI)     │
│  端口: 9092   │  │   端口: 6379       │  │  端口: 3306    │  │   API     │
└──────────────┘  └────────────────────┘  └───────────────┘  └───────────┘
```

---

## 环境要求

### 软件要求

| 软件 | 版本 | 说明 |
|------|------|------|
| Go | 1.21+ | 编译项目 |
| MySQL | 8.0+ | 数据存储 |
| Redis | 7.0+ | 缓存 |
| Kafka | 3.0+ | 消息队列 |
| ZooKeeper | 3.0+ | Kafka 依赖 |
| Node.js | 18+ | 前端开发（可选） |

### 硬件要求

**最低配置**：
- CPU: 2 核
- 内存: 4GB
- 磁盘: 20GB

**推荐配置**：
- CPU: 4 核+
- 内存: 8GB+
- 磁盘: 50GB+

---

## 安装步骤

### 1. 安装依赖服务

#### 使用 Docker 安装（推荐）

```bash
# 启动依赖服务
docker-compose up -d
```

#### 手动安装

**安装 MySQL**:
```bash
# macOS
brew install mysql
brew services start mysql

# Ubuntu/Debian
sudo apt install mysql-server
sudo systemctl start mysql

# 创建数据库
mysql -u root -p -e "CREATE DATABASE goim_chat CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
```

**安装 Redis**:
```bash
# macOS
brew install redis
brew services start redis

# Ubuntu/Debian
sudo apt install redis-server
sudo systemctl start redis
```

**安装 Kafka + ZooKeeper**:
```bash
# 下载 Kafka
wget https://downloads.apache.org/kafka/3.5.1/kafka_2.13-3.5.1.tgz
tar -xzf kafka_2.13-3.5.1.tgz
cd kafka_2.13-3.5.1

# 启动 ZooKeeper
bin/zookeeper-server-start.sh -daemon config/zookeeper.properties

# 启动 Kafka
bin/kafka-server-start.sh -daemon config/server.properties

# 创建主题
bin/kafka-topics.sh --create --topic goim-push-topic --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
```

### 2. 初始化数据库

```bash
# 导入数据库结构
mysql -u root -p goim_chat < scripts/chat-schema.sql

# 导入测试数据（可选）
mysql -u root -p goim_chat < scripts/chat-seed.sql
```

### 3. 编译项目

```bash
# 克隆项目（如果还没有）
git clone https://github.com/your-repo/goim.git
cd goim

# 编译所有组件
go build -o target/comet ./cmd/comet
go build -o target/logic ./cmd/logic
go build -o target/job ./cmd/job
go build -o target/chatapi ./cmd/chatapi
```

### 4. 编译前端（可选）

```bash
cd web
npm install
npm run build
```

---

## 配置说明

### 1. Logic 配置 (`cmd/logic/logic-example.toml`)

```toml
[discovery]
    nodes = ["127.0.0.1:7171"]  # 服务发现节点

[httpServer]
    network = "tcp"
    addr = ":3111"              # HTTP API 端口
    readTimeout = "1s"
    writeTimeout = "1s"

[rpcServer]
    network = "tcp"
    addr = ":3119"              # gRPC 端口
    timeout = "1s"

[kafka]
    topic = "goim-push-topic"
    brokers = ["127.0.0.1:9092"]

[redis]
    network = "tcp"
    addr = "127.0.0.1:6379"
    active = 60000
    idle = 1024
```

### 2. Comet 配置 (`cmd/comet/comet-example.toml`)

```toml
[discovery]
    nodes = ["127.0.0.1:7171"]

[websocket]
    bind = [":3102"]           # WebSocket 端口
    tlsOpen = false            # 是否启用 TLS
    tlsBind = [":3103"]        # TLS WebSocket 端口

[tcp]
    bind = [":3101"]           # TCP 端口

[bucket]
    size = 32                  # Bucket 数量
    channel = 1024             # 每个 Bucket 的 Channel 数
    room = 1024                # 每个 Bucket 的 Room 数
```

### 3. ChatAPI 配置 (`cmd/chatapi/chatapi.toml`)

```toml
[httpServer]
addr = "0.0.0.0:3112"         # HTTP API 端口
readTimeout = 10
writeTimeout = 10

[mysql]
dsn = "root:password@tcp(127.0.0.1:3306)/goim_chat?charset=utf8mb4&parseTime=true&loc=Local"
maxIdle = 10
maxOpen = 100
maxLifetime = 3600

[redis]
network = "tcp"
addr = "127.0.0.1:6379"
auth = ""
database = 0

[logic]
appid = "goim.logic"
timeout = 3000
endpoint = "http://127.0.0.1:3111"

[jwt]
secret = "goim-chat-secret-key-please-change-in-production"
expireTime = 24               # Token 过期时间（小时）

[ai]
provider = "openai"
apiKey = "sk-your-openai-api-key-here"  # 替换为你的 OpenAI API Key
baseUrl = "https://api.openai.com/v1"
model = "gpt-3.5-turbo"
temperature = 0.7
maxTokens = 1000
```

### 4. Job 配置 (`cmd/job/job-example.toml`)

```toml
[discovery]
    nodes = ["127.0.0.1:7171"]

[kafka]
    topic = "goim-push-topic"
    group = "goim-push-group-job"
    brokers = ["127.0.0.1:9092"]
```

---

## 启动服务

### 方式一：手动启动各服务

```bash
# 1. 启动 Logic 服务
./target/logic -conf=cmd/logic/logic-example.toml &

# 2. 启动 Comet 服务
./target/comet -conf=cmd/comet/comet-example.toml &

# 3. 启动 Job 服务
./target/job -conf=cmd/job/job-example.toml &

# 4. 启动 ChatAPI 服务
./target/chatapi cmd/chatapi/chatapi.toml &
```

### 方式二：使用启动脚本

创建 `start.sh`:

```bash
#!/bin/bash

BASE_DIR=$(pwd)
cd $BASE_DIR

echo "Starting GoIM Chat services..."

# 检查服务是否已运行
if [ -f /tmp/goim.pid ]; then
    echo "Services are already running. Use stop.sh first."
    exit 1
fi

# 启动 Logic
echo "Starting Logic..."
nohup $BASE_DIR/target/logic -conf=$BASE_DIR/cmd/logic/logic-example.toml > logs/logic.log 2>&1 &
LOGIC_PID=$!
echo $LOGIC_PID > logs/logic.pid
sleep 2

# 启动 Comet
echo "Starting Comet..."
nohup $BASE_DIR/target/comet -conf=$BASE_DIR/cmd/comet/comet-example.toml > logs/comet.log 2>&1 &
COMET_PID=$!
echo $COMET_PID > logs/comet.pid
sleep 2

# 启动 Job
echo "Starting Job..."
nohup $BASE_DIR/target/job -conf=$BASE_DIR/cmd/job/job-example.toml > logs/job.log 2>&1 &
JOB_PID=$!
echo $JOB_PID > logs/job.pid
sleep 2

# 启动 ChatAPI
echo "Starting ChatAPI..."
nohup $BASE_DIR/target/chatapi $BASE_DIR/cmd/chatapi/chatapi.toml > logs/chatapi.log 2>&1 &
CHATAPI_PID=$!
echo $CHATAPI_PID > logs/chatapi.pid

echo "All services started!"
echo "Logic PID: $LOGIC_PID"
echo "Comet PID: $COMET_PID"
echo "Job PID: $JOB_PID"
echo "ChatAPI PID: $CHATAPI_PID"
```

创建 `stop.sh`:

```bash
#!/bin/bash

echo "Stopping GoIM Chat services..."

# 停止 ChatAPI
if [ -f logs/chatapi.pid ]; then
    kill $(cat logs/chatapi.pid) 2>/dev/null
    rm logs/chatapi.pid
    echo "ChatAPI stopped"
fi

# 停止 Job
if [ -f logs/job.pid ]; then
    kill $(cat logs/job.pid) 2>/dev/null
    rm logs/job.pid
    echo "Job stopped"
fi

# 停止 Comet
if [ -f logs/comet.pid ]; then
    kill $(cat logs/comet.pid) 2>/dev/null
    rm logs/comet.pid
    echo "Comet stopped"
fi

# 停止 Logic
if [ -f logs/logic.pid ]; then
    kill $(cat logs/logic.pid) 2>/dev/null
    rm logs/logic.pid
    echo "Logic stopped"
fi

echo "All services stopped!"
```

赋予执行权限：
```bash
chmod +x start.sh stop.sh
mkdir -p logs
```

### 方式三：使用 systemd（生产环境）

创建 `/etc/systemd/system/goim-chatapi.service`:

```ini
[Unit]
Description=GoIM ChatAPI Service
After=network.target mysql.service redis.service

[Service]
Type=simple
User=goim
WorkingDirectory=/opt/goim
ExecStart=/opt/goim/target/chatapi /opt/goim/cmd/chatapi/chatapi.toml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

启用并启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable goim-chatapi
sudo systemctl start goim-chatapi
```

---

## 测试功能

### 1. 检查服务状态

```bash
# 检查 ChatAPI
curl http://localhost:3112/health

# 检查 Logic
curl http://localhost:3111/health

# 检查端口占用
lsof -i :3101  # Comet TCP
lsof -i :3102  # Comet WebSocket
lsof -i :3111  # Logic HTTP
lsof -i :3112  # ChatAPI
```

### 2. 测试用户注册和登录

```bash
# 注册用户
curl -X POST http://localhost:3112/api/user/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456",
    "nickname": "测试用户"
  }'

# 登录
curl -X POST http://localhost:3112/api/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }'
```

保存返回的 `token` 用于后续测试。

### 3. 测试 AI 聊天

```bash
# 获取 AI Bots 列表
curl -X GET http://localhost:3112/api/ai/bots \
  -H "Authorization: Bearer YOUR_TOKEN"

# 发送消息给 AI
curl -X POST http://localhost:3112/api/ai/chat/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bot_id": 9001,
    "message": "你好，请介绍一下自己"
  }'
```

### 4. 测试 WebSocket 连接

使用浏览器开发者工具或测试页面：

```html
<!DOCTYPE html>
<html>
<head>
    <title>WebSocket Test</title>
</head>
<body>
    <script>
        const ws = new WebSocket('ws://localhost:3102/sub');

        ws.onopen = function() {
            console.log('Connected!');
            // 发送认证消息
            ws.send(JSON.stringify({
                mid: 123,
                key: 'test-key',
                room_id: 'test-room',
                platform: 'web'
            }));
        };

        ws.onmessage = function(event) {
            console.log('Received:', event.data);
        };

        ws.onerror = function(error) {
            console.log('Error:', error);
        };
    </script>
</body>
</html>
```

---

## 常见问题

### 1. 端口被占用

**问题**: `bind: address already in use`

**解决**:
```bash
# 查找占用端口的进程
lsof -i :端口号

# 杀死进程
kill -9 PID

# 或修改配置文件中的端口
```

### 2. MySQL 连接失败

**问题**: `Error 2003: Can't connect to MySQL server`

**解决**:
```bash
# 检查 MySQL 是否运行
mysql -u root -p -e "SELECT 1"

# 检查防火墙
sudo ufw allow 3306

# 检查 MySQL 配置
sudo cat /etc/mysql/my.cnf | grep bind-address
```

### 3. Redis 连接失败

**问题**: `dial tcp 127.0.0.1:6379: connect: connection refused`

**解决**:
```bash
# 检查 Redis 是否运行
redis-cli ping

# 启动 Redis
redis-server --daemonize yes
```

### 4. AI 聊天无响应

**问题**: AI 聊天请求超时

**解决**:
1. 检查 OpenAI API Key 是否正确配置
2. 检查网络是否能访问 OpenAI API
3. 查看日志: `tail -f logs/chatapi.log`

### 5. Kafka 连接失败

**问题**: `kafka: client has run out of available brokers`

**解决**:
```bash
# 检查 Kafka 是否运行
nc -zv localhost 9092

# 检查 ZooKeeper 是否运行
nc -zv localhost 2181

# 重启 Kafka
bin/kafka-server-stop.sh
bin/kafka-server-start.sh config/server.properties
```

### 6. 编译错误

**问题**: `go: module ...: no matching versions`

**解决**:
```bash
# 清理缓存
go clean -modcache

# 更新依赖
go mod tidy
go mod download

# 重新编译
go build -o target/chatapi ./cmd/chatapi
```

---

## 生产环境部署建议

### 1. 安全配置

- 修改 JWT Secret 为强随机字符串
- 使用 HTTPS/WSS 加密传输
- 配置防火墙规则
- 限制 API 访问频率

### 2. 性能优化

- 调整 Bucket、Channel、Room 参数
- 配置 Redis 持久化
- 使用 MySQL 主从复制
- 启用 Nginx 负载均衡

### 3. 监控告警

- 使用 Prometheus + Grafana 监控
- 配置日志收集 (ELK)
- 设置服务健康检查
- 配置告警通知

### 4. 高可用部署

- 部署多个 Comet 实例
- 部署多个 Logic 实例
- 使用 Nginx 做负载均衡
- 配置 Kafka 集群

---

## 参考资料

- [GoIM 官方文档](https://github.com/Terry-Mao/goim)
- [OpenAI API 文档](https://platform.openai.com/docs)
- [Kafka 文档](https://kafka.apache.org/documentation)
- [Redis 文档](https://redis.io/documentation)

---

## 技术支持

如有问题，请提交 Issue 或联系维护者。
