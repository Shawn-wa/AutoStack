.PHONY: all build run dev clean help

# 默认目标
all: help

# ==================== 后端 ====================

# 构建后端（本地macOS）
build-backend:
	@echo "🔨 构建后端..."
	cd backend && go build -o bin/autostack ./cmd/server

# 构建后端（Linux容器用）
build-backend-linux:
	@echo "🔨 构建后端 (Linux ARM64)..."
	cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/autostack ./cmd/server

# 运行后端
run-backend:
	@echo "🚀 启动后端..."
	cd backend && go run cmd/server/main.go

# 后端开发模式（热重载需要 air）
dev-backend:
	@echo "🔄 后端开发模式..."
	cd backend && air

# ==================== 前端 ====================

# 安装前端依赖
install-frontend:
	@echo "📦 安装前端依赖..."
	cd frontend && npm install

# 运行前端开发服务器
run-frontend:
	@echo "🚀 启动前端..."
	cd frontend && npm run dev

# 构建前端
build-frontend:
	@echo "🔨 构建前端..."
	cd frontend && npm run build

# ==================== Docker ====================

# Docker 构建
docker-build:
	@echo "🐳 构建 Docker 镜像..."
	docker-compose build

# Docker 启动
docker-up:
	@echo "🐳 启动 Docker 服务..."
	docker-compose up -d

# Docker 停止
docker-down:
	@echo "🐳 停止 Docker 服务..."
	docker-compose down

# Docker 日志
docker-logs:
	docker-compose logs -f

# ==================== 开发 ====================

# 同时启动前后端（开发模式）
dev:
	@echo "🚀 启动开发环境..."
	@make run-backend &
	@make run-frontend

# 初始化项目
init:
	@echo "📦 初始化项目..."
	cd backend && go mod tidy
	cd frontend && npm install

# ==================== 本地构建 + Docker运行 ====================

# 构建所有（本地构建用于Docker运行）
build:
	@echo "🔨 构建前后端..."
	@make build-backend-linux
	@make build-frontend
	@echo "✅ 构建完成"

# 构建并重启服务
rebuild:
	@echo "🔄 重新构建并重启服务..."
	@make build
	docker-compose restart
	@echo "✅ 服务已重启"

# 快速重启后端（重新构建后端并重启）
rb:
	@echo "🔄 重新构建后端..."
	@make build-backend-linux
	docker-compose restart backend
	@echo "✅ 后端已重启"

# 快速重启前端（重新构建前端并重启）
rf:
	@echo "🔄 重新构建前端..."
	@make build-frontend
	docker-compose restart frontend
	@echo "✅ 前端已重启"

# 清理构建产物
clean:
	@echo "🧹 清理..."
	rm -rf backend/bin
	rm -rf frontend/dist
	rm -rf frontend/node_modules

# ==================== 帮助 ====================

help:
	@echo "AutoStack - 低代码快捷部署平台"
	@echo ""
	@echo "使用方法:"
	@echo "  make init              初始化项目（安装依赖）"
	@echo "  make dev               启动开发环境"
	@echo ""
	@echo "🚀 开发模式（推荐）:"
	@echo "  make build             构建前后端（用于Docker运行）"
	@echo "  make rebuild           构建并重启所有服务"
	@echo "  make rb                快速重启后端（构建+重启）"
	@echo "  make rf                快速重启前端（构建+重启）"
	@echo ""
	@echo "后端:"
	@echo "  make build-backend       构建后端（本地macOS）"
	@echo "  make build-backend-linux 构建后端（Linux容器用）"
	@echo "  make run-backend         运行后端"
	@echo "  make dev-backend         后端开发模式（需要 air）"
	@echo ""
	@echo "前端:"
	@echo "  make install-frontend  安装前端依赖"
	@echo "  make run-frontend      运行前端开发服务器"
	@echo "  make build-frontend    构建前端"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up         启动 Docker 服务"
	@echo "  make docker-down       停止 Docker 服务"
	@echo "  make docker-logs       查看 Docker 日志"
	@echo ""
	@echo "  make clean             清理构建产物"
	@echo "  make help              显示帮助信息"

