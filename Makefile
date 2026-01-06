# GoIM Chat Makefile
# 用于简化项目的编译、测试和部署

.PHONY: all build clean start stop restart test logs help

# Go 参数
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

# 变量定义
BINARY_DIR = target
CMD_DIR = cmd
LOG_DIR = logs
CONFIG_DIR = $(CMD_DIR)

# 默认目标
all: build

## build: 编译所有服务
build:
	@echo "编译 GoIM Chat 服务..."
	@rm -rf $(BINARY_DIR)/
	@mkdir -p $(BINARY_DIR)
	@mkdir -p $(LOG_DIR)
	@cp cmd/comet/comet-example.toml $(BINARY_DIR)/comet.toml
	@cp cmd/logic/logic-example.toml $(BINARY_DIR)/logic.toml
	@cp cmd/job/job-example.toml $(BINARY_DIR)/job.toml
	$(GOBUILD) -o $(BINARY_DIR)/comet ./$(CMD_DIR)/comet
	$(GOBUILD) -o $(BINARY_DIR)/logic ./$(CMD_DIR)/logic
	$(GOBUILD) -o $(BINARY_DIR)/job ./$(CMD_DIR)/job
	$(GOBUILD) -o $(BINARY_DIR)/chatapi ./$(CMD_DIR)/chatapi
	@echo "编译完成！"
	@ls -lh $(BINARY_DIR)/

## clean: 清理编译文件
clean:
	@echo "清理编译文件..."
	@rm -rf $(BINARY_DIR)/*
	@echo "清理完成！"

## start: 启动所有服务
start: build
	@echo "启动 GoIM Chat 服务..."
	@nohup $(BINARY_DIR)/logic -conf=$(BINARY_DIR)/logic.toml > $(LOG_DIR)/logic.log 2>&1 & echo $$! > $(LOG_DIR)/logic.pid
	@sleep 2
	@nohup $(BINARY_DIR)/comet -conf=$(BINARY_DIR)/comet.toml > $(LOG_DIR)/comet.log 2>&1 & echo $$! > $(LOG_DIR)/comet.pid
	@sleep 2
	@nohup $(BINARY_DIR)/job -conf=$(BINARY_DIR)/job.toml > $(LOG_DIR)/job.log 2>&1 & echo $$! > $(LOG_DIR)/job.pid
	@sleep 2
	@nohup $(BINARY_DIR)/chatapi $(CMD_DIR)/chatapi/chatapi.toml > $(LOG_DIR)/chatapi.log 2>&1 & echo $$! > $(LOG_DIR)/chatapi.pid
	@echo "所有服务已启动！使用 'make logs' 查看日志"

## stop: 停止所有服务
stop:
	@echo "停止 GoIM Chat 服务..."
	@if [ -f $(LOG_DIR)/chatapi.pid ]; then kill $$(cat $(LOG_DIR)/chatapi.pid) 2>/dev/null; rm $(LOG_DIR)/chatapi.pid; fi
	@if [ -f $(LOG_DIR)/job.pid ]; then kill $$(cat $(LOG_DIR)/job.pid) 2>/dev/null; rm $(LOG_DIR)/job.pid; fi
	@if [ -f $(LOG_DIR)/comet.pid ]; then kill $$(cat $(LOG_DIR)/comet.pid) 2>/dev/null; rm $(LOG_DIR)/comet.pid; fi
	@if [ -f $(LOG_DIR)/logic.pid ]; then kill $$(cat $(LOG_DIR)/logic.pid) 2>/dev/null; rm $(LOG_DIR)/logic.pid; fi
	@echo "所有服务已停止！"

## restart: 重启所有服务
restart: stop start

## run: 兼容旧的启动命令
run: start

## status: 查看服务状态
status:
	@echo "GoIM Chat 服务状态:"
	@echo "===================="
	@if [ -f $(LOG_DIR)/logic.pid ]; then \
		if ps -p $$(cat $(LOG_DIR)/logic.pid) > /dev/null 2>&1; then \
			echo "Logic:   运行中 (PID: $$(cat $(LOG_DIR)/logic.pid))"; \
		else echo "Logic:   已停止"; fi \
	else echo "Logic:   未启动"; fi
	@if [ -f $(LOG_DIR)/comet.pid ]; then \
		if ps -p $$(cat $(LOG_DIR)/comet.pid) > /dev/null 2>&1; then \
			echo "Comet:   运行中 (PID: $$(cat $(LOG_DIR)/comet.pid))"; \
		else echo "Comet:   已停止"; fi \
	else echo "Comet:   未启动"; fi
	@if [ -f $(LOG_DIR)/job.pid ]; then \
		if ps -p $$(cat $(LOG_DIR)/job.pid) > /dev/null 2>&1; then \
			echo "Job:     运行中 (PID: $$(cat $(LOG_DIR)/job.pid))"; \
		else echo "Job:     已停止"; fi \
	else echo "Job:     未启动"; fi
	@if [ -f $(LOG_DIR)/chatapi.pid ]; then \
		if ps -p $$(cat $(LOG_DIR)/chatapi.pid) > /dev/null 2>&1; then \
			echo "ChatAPI: 运行中 (PID: $$(cat $(LOG_DIR)/chatapi.pid))"; \
		else echo "ChatAPI: 已停止"; fi \
	else echo "ChatAPI: 未启动"; fi
	@echo ""
	@echo "端口占用:"
	@lsof -i :3101 -i :3102 -i :3111 -i :3112 2>/dev/null | grep LISTEN || echo "无服务监听端口"

## logs: 查看日志（ChatAPI）
logs:
	@tail -f $(LOG_DIR)/chatapi.log

## logs-all: 查看所有日志
logs-all:
	@tail -f $(LOG_DIR)/*.log

## test: 运行测试
test:
	@echo "运行测试..."
	$(GOTEST) -v ./...

## test-ai: 测试 AI 聊天功能
test-ai:
	@echo "测试 AI 聊天功能..."
	@curl -s -X POST http://localhost:3112/api/user/login \
		-H "Content-Type: application/json" \
		-d '{"username":"testuser","password":"123456"}' \
		| python3 -m json.tool

## db-init: 初始化数据库
db-init:
	@echo "初始化数据库..."
	@read -p "MySQL 密码: " mysql_pass; \
	mysql -u root -p$$mysql_pass -e "CREATE DATABASE IF NOT EXISTS goim_chat CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"; \
	mysql -u root -p$$mysql_pass goim_chat < scripts/chat-schema.sql; \
	echo "数据库初始化完成！"

## docker-up: 启动 Docker 服务
docker-up:
	@echo "启动 Docker 服务 (Redis, Kafka)..."
	docker-compose up -d

## docker-down: 停止 Docker 服务
docker-down:
	@echo "停止 Docker 服务..."
	docker-compose down

## web-dev: 启动前端开发服务器
web-dev:
	@echo "启动前端开发服务器..."
	cd web && npm run dev

## web-build: 编译前端
web-build:
	@echo "编译前端..."
	cd web && npm run build

## help: 显示帮助信息
help:
	@echo "GoIM Chat Makefile 使用说明"
	@echo "=============================="
	@echo ""
	@echo "编译命令:"
	@echo "  make build           - 编译所有服务"
	@echo "  make clean           - 清理编译文件"
	@echo ""
	@echo "服务管理:"
	@echo "  make start           - 启动所有服务"
	@echo "  make stop            - 停止所有服务"
	@echo "  make restart         - 重启所有服务"
	@echo "  make status          - 查看服务状态"
	@echo ""
	@echo "日志查看:"
	@echo "  make logs            - 查看 ChatAPI 日志"
	@echo "  make logs-all        - 查看所有日志"
	@echo ""
	@echo "测试:"
	@echo "  make test            - 运行测试"
	@echo "  make test-ai         - 测试 AI 聊天"
	@echo ""
	@echo "数据库:"
	@echo "  make db-init         - 初始化数据库"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up       - 启动 Docker 服务"
	@echo "  make docker-down     - 停止 Docker 服务"
	@echo ""
	@echo "前端:"
	@echo "  make web-dev         - 启动前端开发服务器"
	@echo "  make web-build       - 编译前端"
	@echo ""
	@echo "快速启动:"
	@echo "  1. make docker-up    # 启动依赖服务"
	@echo "  2. make db-init      # 初始化数据库"
	@echo "  3. make build        # 编译服务"
	@echo "  4. make start        # 启动服务"
	@echo "  5. make status       # 检查状态"
