#!/bin/bash

set -euo pipefail

usage() {
    cat <<'EOF'
用法: AUTOSTACK_PLATFORM=linux/amd64 ./release.sh <git-tag>

AUTOSTACK_PLATFORM      生产服务器平台，必须指定 linux/amd64 或 linux/arm64
AUTOSTACK_IMAGE_PREFIX  镜像名前缀，默认 ghcr.io/shawn-wa/autostack

已有本地标签时使用该标签的提交，否则从干净的工作区发布 HEAD。
脚本使用 Docker Buildx 构建并推送缺少的镜像，两张镜像就绪后才推送 Git 标签。
运行前请登录目标镜像仓库，并确认有权限向 Git origin 推送标签。
EOF
}

die() {
    printf '[ERROR] %s\n' "$*" >&2
    exit 1
}

if [[ $# -eq 1 && ( $1 == --help || $1 == -h ) ]]; then
    usage
    exit 0
fi
if [[ $# -ne 1 ]]; then
    usage >&2
    exit 1
fi

tag=$1
tag_ref="refs/tags/$tag"
platform=${AUTOSTACK_PLATFORM:-}
image_prefix=${AUTOSTACK_IMAGE_PREFIX:-ghcr.io/shawn-wa/autostack}
case "$platform" in
    linux/amd64|linux/arm64) arch=${platform#linux/} ;;
    *) die '请设置 AUTOSTACK_PLATFORM=linux/amd64 或 AUTOSTACK_PLATFORM=linux/arm64' ;;
esac

cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
git check-ref-format "$tag_ref" && [[ $tag != -* && $tag != HEAD ]] || die "无效的 Git 标签 ${tag}。"
[[ $tag != latest ]] || die 'latest 是部署命令的保留参数，请使用具体版本标签。'

local_tag_object=''
if git show-ref --verify --quiet "$tag_ref"; then
    local_tag_object=$(git rev-parse --verify "$tag_ref")
    commit=$(git rev-parse --verify "$local_tag_object^{commit}") || die "本地标签 $tag 未指向提交"
else
    worktree_status=$(git status --porcelain --untracked-files=normal)
    [[ -z $worktree_status ]] || die '请先提交工作区改动或执行 git stash -u，再从 HEAD 发布新标签'
    commit=$(git rev-parse --verify 'HEAD^{commit}') || die '无法解析 HEAD 提交'
fi

read_remote_tag() {
    local refs
    refs=$(git ls-remote --tags -- "$push_url" "$tag_ref" "$tag_ref^{}") ||
        die '无法读取 origin 的 Git 标签'
    remote_commit=$(awk -v ref="$tag_ref" '
        $2 == ref { object = $1 }
        $2 == ref "^{}" { peeled = $1 }
        END { print (peeled ? peeled : object) }
    ' <<< "$refs")
    [[ -z $remote_commit || $remote_commit == "$commit" ]] ||
        die "origin 的标签 $tag 指向 ${remote_commit}，本次选中的提交为 $commit"
}

push_url=$(git remote get-url --push origin) || die '请配置 Git origin 远程仓库'
read_remote_tag
docker buildx version >/dev/null || die '请安装 Docker 和 Docker Buildx'

printf '发布 %s，提交 %s，平台 %s\n' "$tag" "$commit" "$platform"
source_dir=''
trap 'if [[ -n $source_dir ]]; then rm -rf -- "$source_dir"; fi' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

publish_image() {
    local service=$1
    local image="${image_prefix}-${service}:${commit}-${arch}"

    if docker buildx imagetools inspect -- "$image" >/dev/null 2>&1; then
        printf '复用已发布镜像 %s\n' "$image"
        return
    fi

    if [[ -z $source_dir ]]; then
        source_dir=$(mktemp -d "${TMPDIR:-/tmp}/autostack-release.XXXXXXXX")
        git archive --format=tar "$commit" backend frontend | tar -xf - -C "$source_dir" ||
            die "无法导出提交 $commit 的构建文件"
    fi

    printf '构建并推送镜像 %s\n' "$image"
    docker buildx build --platform "$platform" --tag "$image" \
        --label "org.opencontainers.image.revision=$commit" --push -- "$source_dir/$service" ||
        die "镜像发布失败 ${image}。解决错误后可重试发布。"
}

publish_image backend
publish_image frontend

read_remote_tag
if [[ -z $remote_commit ]]; then
    if [[ -z $local_tag_object ]]; then
        git tag --no-sign -- "$tag" "$commit" || die "无法创建标签 $tag"
        local_tag_object=$commit
    elif [[ $(git rev-parse --verify "$tag_ref") != "$local_tag_object" ]]; then
        die "本地标签 $tag 在发布期间发生变化"
    fi
    git push --no-follow-tags origin "$local_tag_object:$tag_ref" ||
        die "无法推送标签 ${tag}。镜像已就绪，解决错误后可重试发布。"
fi

printf '版本 %s 已就绪\n' "$tag"
