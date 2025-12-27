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

# 4. 启动后端和前端容器（使用旧镜像）
echo "🔄 启动服务容器..."
docker-compose up -d backend frontend

# 5. 复制最新的后端代码到容器
echo "📤 更新后端代码..."
docker cp backend/bin/autostack autostack-backend:/app/autostack

# 6. 复制最新的前端代码到容器
echo "📤 更新前端代码..."
docker cp frontend/dist/. autostack-frontend:/usr/share/nginx/html/

# 7. 重启服务
echo "🔄 重启服务..."
docker restart autostack-backend
docker exec autostack-frontend nginx -s reload

echo ""
echo "✅ 部署完成！"
echo "   前端: http://localhost:3000"
echo "   后端: http://localhost:8080"

