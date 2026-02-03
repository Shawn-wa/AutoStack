#!/bin/bash

# AutoStack 生产环境一键部署脚本
# 使用方法: ./deploy.sh [start|stop|restart|logs|status|update]

set -e

COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Docker 是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
}

# 检查环境变量文件
check_env() {
    if [ ! -f "$ENV_FILE" ]; then
        if [ -f ".env.production" ]; then
            log_warn "未找到 .env 文件，从 .env.production 复制"
            cp .env.production .env
        else
            log_error "未找到环境变量配置文件"
            exit 1
        fi
    fi
}

# Docker Compose 命令（兼容新旧版本）
docker_compose() {
    if docker compose version &> /dev/null; then
        docker compose -f "$COMPOSE_FILE" "$@"
    else
        docker-compose -f "$COMPOSE_FILE" "$@"
    fi
}

# 启动服务
start() {
    log_info "正在启动 AutoStack 服务..."
    check_docker
    check_env
    docker_compose up -d --build
    log_info "服务启动完成！"
    echo ""
    log_info "访问地址: http://$(hostname -I | awk '{print $1}' 2>/dev/null || echo 'localhost')"
    echo ""
    status
}

# 停止服务
stop() {
    log_info "正在停止 AutoStack 服务..."
    docker_compose down
    log_info "服务已停止"
}

# 重启服务
restart() {
    log_info "正在重启 AutoStack 服务..."
    docker_compose restart
    log_info "服务重启完成"
    status
}

# 查看日志
logs() {
    SERVICE=${1:-""}
    if [ -n "$SERVICE" ]; then
        docker_compose logs -f "$SERVICE"
    else
        docker_compose logs -f
    fi
}

# 查看状态
status() {
    log_info "服务状态:"
    docker_compose ps
}

# 更新部署
update() {
    log_info "正在更新 AutoStack..."
    git pull
    docker_compose up -d --build
    log_info "更新完成！"
    status
}

# 显示帮助
show_help() {
    echo "AutoStack 部署脚本"
    echo ""
    echo "使用方法: ./deploy.sh [命令]"
    echo ""
    echo "命令:"
    echo "  start     启动服务（首次会构建镜像）"
    echo "  stop      停止服务"
    echo "  restart   重启服务"
    echo "  logs      查看日志（可指定服务：logs backend）"
    echo "  status    查看服务状态"
    echo "  update    拉取代码并重新部署"
    echo "  help      显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  ./deploy.sh start        # 启动所有服务"
    echo "  ./deploy.sh logs backend # 查看后端日志"
}

# 主入口
case "${1:-start}" in
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    logs)
        logs "$2"
        ;;
    status)
        status
        ;;
    update)
        update
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "未知命令: $1"
        show_help
        exit 1
        ;;
esac
