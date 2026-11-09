#!/bin/bash

set -e -o pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
COMPOSE_FILE=docker-compose.prod.yml
ENV_FILE=.env
STATE_FILE=.deploy_state
LOCK_DIR=.deploy.lock
WAIT_TIMEOUT=${AUTOSTACK_WAIT_TIMEOUT:-120}
OFFLINE=${AUTOSTACK_OFFLINE:-0}
IMAGE_TAG=
STATE_TMP=

log_info() {
    printf '[INFO] %s\n' "$*"
}

log_warn() {
    printf '[WARN] %s\n' "$*" >&2
}

log_error() {
    printf '[ERROR] %s\n' "$*" >&2
}

die() {
    log_error "$*"
    exit 1
}

check_docker() {
    if ! command -v docker >/dev/null 2>&1 || ! docker compose version >/dev/null 2>&1; then
        die "请安装 Docker Engine 和 Docker Compose 插件（docker compose）。"
    fi
}

check_release_compose() {
    local up_help pull_help
    up_help=$(docker compose up --help)
    pull_help=$(docker compose pull --help)
    if [[ $up_help != *--wait-timeout* || $up_help != *--pull* || $pull_help != *--policy* ]]; then
        die "请升级 Docker Compose 插件，部署需要 up --wait-timeout、up --pull 和 pull --policy。"
    fi
    if [[ ! $WAIT_TIMEOUT =~ ^[1-9][0-9]*$ ]]; then
        die "AUTOSTACK_WAIT_TIMEOUT 必须是正整数秒。"
    fi
}

check_env() {
    if [[ ! -f $ENV_FILE ]]; then
        if [[ -f .env.production ]]; then
            log_warn "未找到 .env，从 .env.production 复制。"
            cp .env.production "$ENV_FILE"
        else
            die "未找到 .env，请先配置服务器环境变量。"
        fi
    fi
}

docker_compose() {
    local files=(-f "$COMPOSE_FILE")
    if [[ -n ${AUTOSTACK_COMPOSE_OVERRIDE:-} ]]; then
        files+=(-f "$AUTOSTACK_COMPOSE_OVERRIDE")
    fi
    AUTOSTACK_IMAGE_TAG=${IMAGE_TAG:-unselected} docker compose "${files[@]}" "$@"
}

cleanup() {
    if [[ -n $STATE_TMP ]]; then
        rm -f -- "$STATE_TMP"
    fi
    rmdir -- "$LOCK_DIR"
}

lock_deploy() {
    if ! mkdir -- "$LOCK_DIR" 2>/dev/null; then
        die "无法获取部署锁 ${LOCK_DIR}，请检查是否已有部署或服务管理命令运行。"
    fi
    trap cleanup EXIT
    trap 'exit 130' INT
    trap 'exit 143' TERM
    trap 'exit 129' HUP
}

read_state() {
    CURRENT=
    CURRENT_IMAGE=
    PREVIOUS=
    PREVIOUS_IMAGE=
    if [[ -f $STATE_FILE ]]; then
        # 当前版本和上一版本各占一行：Git tag、镜像 tag。
        {
            IFS=$'\t' read -r CURRENT CURRENT_IMAGE
            IFS=$'\t' read -r PREVIOUS PREVIOUS_IMAGE
        } < "$STATE_FILE"
    else
        if [[ -f .deployed_version ]]; then
            CURRENT=$(<.deployed_version)
        fi
        if [[ -f .deployed_version.previous ]]; then
            PREVIOUS=$(<.deployed_version.previous)
        fi
    fi
}

remote_tags() {
    log_info "正在查询远程 tags..." >&2
    git ls-remote --tags --refs --sort=-version:refname origin |
        cut -f2 | sed 's|^refs/tags/||'
}

list_tags() {
    remote_tags | sed -n '1,20p'
    read_state
    if [[ -n $CURRENT ]]; then
        log_info "当前部署版本：$CURRENT"
    fi
}

show_version() {
    read_state
    if [[ -n $CURRENT ]]; then
        log_info "当前部署版本：$CURRENT"
        if [[ -n $CURRENT_IMAGE ]]; then
            log_info "当前镜像标签：$CURRENT_IMAGE"
        fi
    else
        log_warn "尚无部署版本记录。"
    fi
    if [[ -n $PREVIOUS ]]; then
        log_info "上一部署版本：$PREVIOUS"
    fi
}

resolve_image_tag() {
    local tag=$1 ref commit platform
    ref="refs/tags/$tag"
    git check-ref-format "$ref" || die "无效的 Git tag：$tag"
    if ! commit=$(git rev-parse --verify --quiet "${ref}^{commit}"); then
        if [[ $OFFLINE == 1 ]]; then
            die "本地缺少 Git tag ${tag}，请先联网运行 ./deploy.sh prepare ${tag}。"
        fi
        log_info "正在获取 Git tag $tag..."
        git fetch --no-tags --no-recurse-submodules --depth 1 origin "$ref:$ref" ||
            die "无法从 origin 获取 Git tag ${tag}。"
        commit=$(git rev-parse --verify --quiet "${ref}^{commit}") ||
            die "Git tag $tag 未指向提交。"
    fi
    platform=$(docker version --format '{{.Server.Os}}/{{.Server.Arch}}') ||
        die "无法读取 Docker Engine 的平台，请检查 Docker 是否运行。"
    case "$platform" in
        linux/amd64|linux/arm64)
            IMAGE_TAG="$commit-${platform#*/}"
            ;;
        *)
            die "不支持的 Docker Engine 平台：${platform}（支持 linux/amd64 和 linux/arm64）。"
            ;;
    esac
}

pull_app_images() {
    log_info "正在准备前后端镜像 $IMAGE_TAG..."
    docker_compose pull --policy missing backend frontend ||
        die "无法准备前后端镜像。请检查对应版本是否已发布，以及镜像仓库登录和网络。"
}

cached_images() {
    local images image
    images=$(docker_compose config --images "$@") || return 1
    [[ -n $images ]] || return 1
    while IFS= read -r image; do
        if ! docker image inspect -- "$image" >/dev/null 2>&1; then
            log_warn "本地缺少 Docker 镜像：$image"
            return 1
        fi
    done <<< "$images"
}

stage_release() {
    if [[ $OFFLINE == 1 ]]; then
        cached_images || die "离线启动需要本地镜像，请先联网准备前后端和 MySQL 镜像。"
    else
        pull_app_images
        if ! cached_images mysql 2>/dev/null; then
            log_info "正在准备 MySQL 镜像..."
            docker_compose pull --policy missing mysql ||
                die "无法准备 MySQL 镜像。"
        fi
    fi
}

up_healthy() {
    docker_compose up -d --no-build --pull never --wait --wait-timeout "$WAIT_TIMEOUT" "$@"
}

restore_current() {
    if [[ -z $CURRENT_IMAGE || $CURRENT_IMAGE == "$IMAGE_TAG" ]]; then
        return 1
    fi
    IMAGE_TAG=$CURRENT_IMAGE
    log_warn "正在使用本地镜像恢复 $CURRENT..."
    if cached_images backend frontend && up_healthy --force-recreate backend frontend; then
        log_info "已恢复到 ${CURRENT}。"
    else
        return 1
    fi
}

activate_release() {
    local tag=$1 prior_tag=$PREVIOUS prior_image=$PREVIOUS_IMAGE current_commit
    stage_release

    # 迁移时可用预先缓存的旧版本镜像恢复，无需为恢复再次联网。
    if [[ -n $CURRENT && -z $CURRENT_IMAGE ]] &&
        current_commit=$(git rev-parse --verify --quiet "refs/tags/$CURRENT^{commit}"); then
        CURRENT_IMAGE="$current_commit-${IMAGE_TAG##*-}"
    fi
    if [[ -n $CURRENT && ( $tag != "$CURRENT" || ( -n $CURRENT_IMAGE && $IMAGE_TAG != "$CURRENT_IMAGE" ) ) ]]; then
        prior_tag=$CURRENT
        prior_image=$CURRENT_IMAGE
    fi
    STATE_TMP=$(mktemp "$STATE_FILE.tmp.XXXXXX")
    printf '%s\t%s\n%s\t%s\n' "$tag" "$IMAGE_TAG" "$prior_tag" "$prior_image" > "$STATE_TMP"

    log_info "正在启动版本 $tag..."
    if ! up_healthy; then
        log_error "版本 $tag 未能通过启动检查。"
    elif ! mv -f -- "$STATE_TMP" "$STATE_FILE"; then
        log_error "无法保存版本记录 ${STATE_FILE}。"
    else
        STATE_TMP=
        log_info "部署完成：${tag}（${IMAGE_TAG}）"
        return 0
    fi

    if ! restore_current; then
        log_warn "未能自动恢复，请查看 ./deploy.sh status 和 ./deploy.sh logs [service]。修复后可重新部署同一 tag。"
    fi
    return 1
}

require_tag() {
    [[ -n $1 ]] || die "请指定 Git tag，例如 ./deploy.sh deploy v1.0.0。"
}

prepare_tag() {
    require_tag "$1"
    [[ $OFFLINE != 1 ]] || die "准备镜像需要联网，请取消 AUTOSTACK_OFFLINE。"
    resolve_image_tag "$1"
    pull_app_images
    log_info "已缓存版本 $1 的前后端镜像。"
}

deploy_tag() {
    require_tag "$1"
    resolve_image_tag "$1"
    activate_release "$1"
}

deploy_latest() {
    local latest
    [[ $OFFLINE != 1 ]] || die "查询远程最新版本需要联网。离线部署请指定 Git tag。"
    latest=$(remote_tags | sed -n '1p')
    [[ -n $latest ]] || die "没有找到任何 Git tag。"
    deploy_tag "$latest"
}

start() {
    [[ -n $CURRENT ]] || die "尚无部署版本记录，请先运行 ./deploy.sh deploy <tag>。"
    IMAGE_TAG=$CURRENT_IMAGE
    if [[ -z $IMAGE_TAG ]]; then
        resolve_image_tag "$CURRENT"
    fi
    activate_release "$CURRENT"
}

rollback() {
    [[ -n $PREVIOUS ]] || die "没有找到上一个版本记录。"
    log_warn "准备从 $CURRENT 回滚到 ${PREVIOUS}。"
    local reply
    read -r -p "确认回滚? (y/N) " reply || reply=
    if [[ $reply != [Yy] ]]; then
        log_info "已取消回滚。"
        return 0
    fi
    IMAGE_TAG=$PREVIOUS_IMAGE
    if [[ -z $IMAGE_TAG ]]; then
        resolve_image_tag "$PREVIOUS"
    fi
    activate_release "$PREVIOUS"
}

show_help() {
    cat <<'HELP'
AutoStack 生产部署

  ./deploy.sh tags             列出可用的 Git tags
  ./deploy.sh version          显示当前和上一次成功部署
  ./deploy.sh prepare <tag>     提前拉取前后端镜像
  ./deploy.sh deploy <tag>      启动指定版本
  ./deploy.sh deploy latest     启动按版本排序的最新 tag
  ./deploy.sh update            同 deploy latest
  ./deploy.sh rollback          回滚到上一个成功版本
  ./deploy.sh start             启动当前记录的版本
  ./deploy.sh stop              停止服务
  ./deploy.sh restart           重启容器
  ./deploy.sh status            查看服务状态
  ./deploy.sh logs [service]    查看服务日志

AUTOSTACK_OFFLINE=1 使用已经缓存的 Git tag 和 Docker 镜像。
AUTOSTACK_COMPOSE_OVERRIDE 可指定额外的 Compose 文件（例如 docker-compose.ssl.yml）。
AUTOSTACK_WAIT_TIMEOUT 可设置 Compose 健康检查等待秒数（默认 120）。
HELP
}

command=${1:-help}
case "$command" in
    prepare|deploy) require_tag "${2:-}" ;;
esac
case "$command" in
    prepare|deploy|update|start|rollback|stop|restart)
        check_docker
        lock_deploy
        read_state
        IMAGE_TAG=$CURRENT_IMAGE
        case "$command" in
            prepare|deploy|update|start|rollback)
                check_env
                check_release_compose
                ;;
        esac
        ;;
    status|logs)
        check_docker
        read_state
        IMAGE_TAG=$CURRENT_IMAGE
        ;;
esac

case "$command" in
    prepare) prepare_tag "${2:-}" ;;
    deploy)
        if [[ ${2:-} == latest ]]; then
            deploy_latest
        else
            deploy_tag "${2:-}"
        fi
        ;;
    update) deploy_latest ;;
    start) start ;;
    rollback) rollback ;;
    stop) docker_compose down ;;
    restart) docker_compose restart ;;
    status) docker_compose ps ;;
    logs)
        if [[ -n ${2:-} ]]; then
            docker_compose logs -f -- "$2"
        else
            docker_compose logs -f
        fi
        ;;
    tags) list_tags ;;
    version) show_version ;;
    help|--help|-h) show_help ;;
    *)
        log_error "未知命令：$command"
        show_help
        exit 1
        ;;
esac
