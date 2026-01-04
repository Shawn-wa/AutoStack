#!/bin/bash
# AutoStack 开发模式启动脚本

set -e

echo "🔧 AutoStack 开发模式"
echo "===================="

# 检查参数
case "$1" in
  "mysql")
    echo "🗄️  启动 MySQL..."
    docker-compose up -d mysql
    echo "⏳ 等待 MySQL 就绪..."
    sleep 5
    echo "✅ MySQL 已启动 (端口 3306)"
    ;;
  "frontend")
    echo "🎨 启动前端开发服务器..."
    cd frontend
    npm run dev
    ;;
  "backend")
    echo "⚙️  启动后端开发服务器..."
    cd backend
    # 设置环境变量连接 Docker MySQL
    export DATABASE_DRIVER=mysql
    export DATABASE_DSN="root:As2024#xK9mPv7Rn@tcp(localhost:3306)/autostack?charset=utf8mb4&parseTime=True&loc=Local"
    export SERVER_MODE=debug
    go run ./cmd/server
    ;;
  "all")
    echo "🚀 启动所有服务..."
    # 启动 MySQL
    docker-compose up -d mysql
    echo "⏳ 等待 MySQL 就绪..."
    sleep 5
    
    # 后台启动后端
    echo "⚙️  启动后端..."
    cd backend
    export DATABASE_DRIVER=mysql
    export DATABASE_DSN="root:As2024#xK9mPv7Rn@tcp(localhost:3306)/autostack?charset=utf8mb4&parseTime=True&loc=Local"
    export SERVER_MODE=debug
    go run ./cmd/server &
    BACKEND_PID=$!
    cd ..
    
    # 等待后端启动
    sleep 3
    
    # 启动前端
    echo "🎨 启动前端..."
    cd frontend
    npm run dev &
    FRONTEND_PID=$!
    cd ..
    
    echo ""
    echo "✅ 开发环境已启动！"
    echo "   前端: http://localhost:3000 (热重载)"
    echo "   后端: http://localhost:8080"
    echo ""
    echo "按 Ctrl+C 停止所有服务"
    
    # 捕获退出信号
    trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null" EXIT
    wait
    ;;
  *)
    echo "用法: ./dev.sh [命令]"
    echo ""
    echo "命令:"
    echo "  mysql     只启动 MySQL 容器"
    echo "  frontend  启动前端开发服务器 (热重载)"
    echo "  backend   启动后端开发服务器"
    echo "  all       启动所有服务"
    echo ""
    echo "推荐开发流程（分开3个终端）:"
    echo "  终端1: ./dev.sh mysql"
    echo "  终端2: ./dev.sh backend"
    echo "  终端3: ./dev.sh frontend"
    echo ""
    echo "前端修改会自动热重载，无需重启！"
    ;;
esac

