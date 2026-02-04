#!/bin/bash

# AutoStack 生产环境一键部署脚本
# 使用方法: ./deploy.sh [命令] [参数]

set -e

COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env"
VERSION_FILE=".deployed_version"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_version() {
    echo -e "${BLUE}[VERSION]${NC} $1"
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

# 获取远程 tags
fetch_tags() {
    log_info "正在获取远程 tags..."
    git fetch --tags --force
}

# 列出所有可用的 tags
list_tags() {
    fetch_tags
    echo ""
    log_info "可用版本列表（按时间倒序）："
    echo ""
    git tag -l --sort=-version:refname | head -20
    echo ""
    CURRENT=$(get_current_version)
    if [ -n "$CURRENT" ]; then
        log_version "当前部署版本: $CURRENT"
    fi
}

# 获取当前部署的版本
get_current_version() {
    if [ -f "$VERSION_FILE" ]; then
        cat "$VERSION_FILE"
    else
        # 尝试从 git 获取当前 tag
        git describe --tags --exact-match 2>/dev/null || echo ""
    fi
}

# 获取最新的 tag
get_latest_tag() {
    git tag -l --sort=-version:refname | head -1
}

# 显示当前版本
show_version() {
    CURRENT=$(get_current_version)
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    
    echo ""
    echo "AutoStack 版本信息"
    echo "─────────────────────────────"
    if [ -n "$CURRENT" ]; then
        log_version "部署版本: $CURRENT"
    else
        log_warn "部署版本: 未记录（可能是直接从分支部署）"
    fi
    echo "Git 分支:  $GIT_BRANCH"
    echo "Git 提交:  $GIT_COMMIT"
    echo ""
}

# 部署指定 tag 版本
deploy_tag() {
    TAG=$1
    
    if [ -z "$TAG" ]; then
        log_error "请指定要部署的版本号"
        echo ""
        echo "用法: ./deploy.sh deploy <tag>"
        echo "示例: ./deploy.sh deploy v1.0.0"
        echo ""
        echo "查看可用版本: ./deploy.sh tags"
        exit 1
    fi
    
    fetch_tags
    
    # 检查 tag 是否存在
    if ! git rev-parse "$TAG" &>/dev/null; then
        log_error "版本 $TAG 不存在"
        echo ""
        echo "查看可用版本: ./deploy.sh tags"
        exit 1
    fi
    
    CURRENT=$(get_current_version)
    if [ "$CURRENT" = "$TAG" ]; then
        log_warn "当前已是版本 $TAG，无需重新部署"
        exit 0
    fi
    
    log_info "准备部署版本: $TAG"
    if [ -n "$CURRENT" ]; then
        log_info "当前版本: $CURRENT"
    fi
    echo ""
    
    # 保存当前版本用于回滚
    if [ -n "$CURRENT" ]; then
        echo "$CURRENT" > "${VERSION_FILE}.previous"
    fi
    
    # 切换到指定 tag
    log_info "切换到版本 $TAG..."
    git checkout "$TAG"
    
    # 记录部署版本
    echo "$TAG" > "$VERSION_FILE"
    
    # 重新构建并启动
    check_docker
    check_env
    log_info "正在构建并启动服务..."
    docker_compose up -d --build
    
    echo ""
    log_info "部署完成！"
    log_version "当前版本: $TAG"
    echo ""
    status
}

# 部署最新 tag
deploy_latest() {
    fetch_tags
    LATEST=$(get_latest_tag)
    
    if [ -z "$LATEST" ]; then
        log_error "没有找到任何 tag，请先创建版本标签"
        echo ""
        echo "创建 tag 示例: git tag v1.0.0 && git push origin v1.0.0"
        exit 1
    fi
    
    log_info "最新版本: $LATEST"
    deploy_tag "$LATEST"
}

# 回滚到上一个版本
rollback() {
    if [ ! -f "${VERSION_FILE}.previous" ]; then
        log_error "没有找到上一个版本记录，无法回滚"
        exit 1
    fi
    
    PREVIOUS=$(cat "${VERSION_FILE}.previous")
    CURRENT=$(get_current_version)
    
    log_warn "准备回滚"
    log_info "当前版本: $CURRENT"
    log_info "回滚到:   $PREVIOUS"
    echo ""
    
    read -p "确认回滚? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "已取消回滚"
        exit 0
    fi
    
    deploy_tag "$PREVIOUS"
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

# 更新部署（拉取最新 tag）
update() {
    log_info "正在更新到最新版本..."
    deploy_latest
}

# 显示帮助
show_help() {
    echo "AutoStack 部署脚本"
    echo ""
    echo "使用方法: ./deploy.sh [命令] [参数]"
    echo ""
    echo "版本管理命令:"
    echo "  tags              列出所有可用版本（git tags）"
    echo "  version           显示当前部署的版本"
    echo "  deploy <tag>      部署指定版本（如: deploy v1.0.0）"
    echo "  deploy latest     部署最新版本"
    echo "  update            更新到最新版本（等同于 deploy latest）"
    echo "  rollback          回滚到上一个版本"
    echo ""
    echo "服务管理命令:"
    echo "  start             启动服务（首次会构建镜像）"
    echo "  stop              停止服务"
    echo "  restart           重启服务"
    echo "  status            查看服务状态"
    echo "  logs [service]    查看日志（可指定服务：logs backend）"
    echo ""
    echo "示例:"
    echo "  ./deploy.sh tags              # 查看所有可用版本"
    echo "  ./deploy.sh deploy v1.2.0     # 部署 v1.2.0 版本"
    echo "  ./deploy.sh deploy latest     # 部署最新版本"
    echo "  ./deploy.sh rollback          # 回滚到上一版本"
    echo "  ./deploy.sh logs backend      # 查看后端日志"
}

# 主入口
case "${1:-help}" in
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
    tags)
        list_tags
        ;;
    version)
        show_version
        ;;
    deploy)
        if [ "$2" = "latest" ]; then
            deploy_latest
        else
            deploy_tag "$2"
        fi
        ;;
    update)
        update
        ;;
    rollback)
        rollback
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
