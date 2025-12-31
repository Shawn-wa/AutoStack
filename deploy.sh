#!/bin/bash
# AutoStack 本地部署脚本

set -e

echo "🚀 AutoStack 部署脚本"
echo "===================="

# 1. 构建后端（Linux版本）
echo "📦 构建后端..."
cd backend
GOOS=linux GOARCH=amd64 go build -o bin/autostack ./cmd/server
cd ..

# 2. 构建前端
echo "📦 构建前端..."
cd frontend
npm run build
cd ..

# 3. 启动MySQL容器（如果未运行）
echo "🗄️  启动MySQL..."
docker-compose up -d mysql
echo "⏳ 等待MySQL就绪..."
sleep 10

# 4. 启动后端和前端容器（使用volume映射，自动加载最新构建）
echo "🔄 启动服务容器..."
docker-compose up -d backend frontend

# 5. 重启后端服务加载新二进制
echo "🔄 重启后端..."
docker restart autostack-backend

# 6. 重载nginx配置
echo "🔄 重载前端..."
docker exec autostack-frontend nginx -s reload

echo ""
echo "✅ 部署完成！"
echo "   前端: http://localhost:3000"
echo "   后端: http://localhost:8080"

