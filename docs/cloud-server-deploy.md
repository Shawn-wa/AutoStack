# 云服务器部署指南

本文档介绍如何将 AutoStack 部署到云服务器（阿里云、腾讯云、AWS 等）。

## 一、前置条件

### 1.1 服务器要求

| 配置项 | 最低要求 | 推荐配置 |
|--------|---------|---------|
| CPU | 1 核 | 2 核+ |
| 内存 | 2 GB | 4 GB+ |
| 硬盘 | 20 GB | 40 GB+ SSD |
| 系统 | Ubuntu 20.04+ / CentOS 7+ | Ubuntu 22.04 LTS |
| 带宽 | 1 Mbps | 5 Mbps+ |

### 1.2 开放端口

在云服务器安全组中开放以下端口：

| 端口 | 用途 | 必须 |
|------|------|------|
| 22 | SSH 远程连接 | 是 |
| 80 | HTTP 访问 | 是 |
| 443 | HTTPS 访问 | 推荐 |

> ⚠️ **安全提示**：MySQL 3306 端口不要对外开放，保持仅内部访问。

## 二、服务器环境准备

### 2.1 安装 Docker

**Ubuntu/Debian:**

```bash
# 更新包索引
sudo apt update

# 安装依赖
sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release

# 添加 Docker GPG 密钥
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# 添加 Docker 仓库
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装 Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动 Docker
sudo systemctl enable docker
sudo systemctl start docker

# 将当前用户加入 docker 组（免 sudo）
sudo usermod -aG docker $USER
```

**CentOS/RHEL:**

```bash
# 安装依赖
sudo yum install -y yum-utils

# 添加 Docker 仓库
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# 安装 Docker
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动 Docker
sudo systemctl enable docker
sudo systemctl start docker

# 将当前用户加入 docker 组
sudo usermod -aG docker $USER
```

### 2.2 安装 Git

```bash
# Ubuntu/Debian
sudo apt install -y git

# CentOS
sudo yum install -y git
```

### 2.3 验证安装

```bash
docker --version
docker compose version
git --version
# 记录 Docker Engine 的目标架构，用于开发机打包
docker version --format '{{.Server.Os}}/{{.Server.Arch}}'
```

生产脚本使用支持 `pull --policy` 和 `up --wait` 的 Docker Compose V2。

## 三、部署项目

### 3.1 克隆代码

```bash
# 进入部署目录
cd /opt

# 克隆项目（替换为你的仓库地址）
sudo git clone https://github.com/your-username/AutoStack.git
cd AutoStack

# 修改目录权限
sudo chown -R $USER:$USER /opt/AutoStack
```

### 3.2 配置环境变量

升级已有部署时，沿用原目录的 `.env` 和 `./data/mysql`。已有平台凭据需要原来的 `CRYPTO_SECRET_KEY` 才能解密。

```bash
# 复制环境变量模板
cp .env.production .env

# 编辑配置（修改为安全的密钥）
vim .env
```

**建议修改以下配置：**

```bash
# MySQL 密码（使用强密码）
MYSQL_PASSWORD=你的强密码

# JWT 密钥（重新生成）
JWT_SECRET=你的JWT密钥

# 加密密钥（重新生成）
CRYPTO_SECRET_KEY=你的加密密钥
```

**生成安全密钥的方法：**

```bash
# 生成 JWT 密钥
openssl rand -base64 32

# 生成加密密钥（32字符）
openssl rand -base64 24
```

### 3.3 部署首个版本

先在开发机[发布镜像和 Git Tag](#52-版本发布流程开发端)。默认镜像名为 `ghcr.io/shawn-wa/autostack-backend` 和 `ghcr.io/shawn-wa/autostack-frontend`。使用其他仓库时，在开发机导出 `AUTOSTACK_IMAGE_PREFIX`，在服务器 `.env` 中填写同一个前缀，例如 `registry.example.com/team/autostack`。

生产机访问私有 GHCR 镜像时，用有镜像读取权限的账号和 `read:packages` 令牌登录。 [GHCR 认证](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry)

```bash
docker login ghcr.io
./deploy.sh deploy v1.2.0
```

用实际已发布的 Tag 替换示例版本。首次部署还会启动 MySQL 并等待初始化完成。

### 3.4 验证部署

```bash
# 查看服务状态
./deploy.sh status

# 查看日志
./deploy.sh logs
```

所有服务状态应为 `Up`：

```
NAME                 STATUS
autostack-mysql      Up (healthy)
autostack-backend    Up (healthy)
autostack-frontend   Up (healthy)
```

### 3.5 访问测试

在浏览器访问：`http://你的服务器公网IP`

## 四、域名配置（可选）

### 4.1 域名解析

在你的域名服务商处添加 A 记录：

| 主机记录 | 记录类型 | 记录值 |
|---------|---------|--------|
| @ 或 www | A | 服务器公网 IP |

### 4.2 配置 HTTPS（推荐）

使用 Let's Encrypt 免费证书，通过 Certbot 自动申请和续期。

**方法一：使用独立的 Nginx + Certbot**

```bash
# 停止现有服务
./deploy.sh stop

# 安装 Certbot
sudo apt install -y certbot

# 申请证书（standalone 模式）
sudo certbot certonly --standalone -d your-domain.com

# 证书位置
# /etc/letsencrypt/live/your-domain.com/fullchain.pem
# /etc/letsencrypt/live/your-domain.com/privkey.pem
```

**方法二：使用 docker-compose 集成 HTTPS**

创建 `docker-compose.ssl.yml`：

```yaml
version: '3.8'

services:
  frontend:
    volumes:
      - /etc/letsencrypt:/etc/letsencrypt:ro
      - ./nginx-ssl.conf:/etc/nginx/conf.d/default.conf:ro
    ports:
      - "80:80"
      - "443:443"
```

创建 `nginx-ssl.conf`：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /etc/letsencrypt/live/your-domain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/your-domain.com/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256;
    ssl_prefer_server_ciphers off;

    root /usr/share/nginx/html;
    index index.html;

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

使用此配置时，在运维账号环境中设置 `AUTOSTACK_COMPOSE_OVERRIDE`，让后续部署也加载同一份覆盖文件。

```bash
export AUTOSTACK_COMPOSE_OVERRIDE=docker-compose.ssl.yml
./deploy.sh start
```

### 4.3 证书自动续期

```bash
# 测试续期
sudo certbot renew --dry-run

# 添加定时任务（每天检查）
echo "0 3 * * * root certbot renew --quiet && docker compose -f /opt/AutoStack/docker-compose.prod.yml restart frontend" | sudo tee /etc/cron.d/certbot-renew
```

## 五、版本管理与更新

开发机先构建并推送前后端成品镜像，随后发布 Git Tag。服务器根据 Tag 对应的 Git 提交和 Docker Engine 架构选择镜像，在现有 Compose 项目中更新容器。

### 5.1 版本管理命令

```bash
# 查看所有可用版本
./deploy.sh tags

# 查看当前部署的版本
./deploy.sh version

# 提前下载镜像，旧容器继续运行
./deploy.sh prepare v1.2.0

# 部署指定版本
./deploy.sh deploy v1.2.0

# 部署最新版本
./deploy.sh deploy latest
# 或
./deploy.sh update

# 回滚到上一个版本
./deploy.sh rollback
```

### 5.2 版本发布流程（开发端）

在开发机提交本次发版所需的改动。根据服务器 Docker Engine 的输出设置 `AUTOSTACK_PLATFORM`，例如 `linux/amd64` 或 `linux/arm64`。使用默认 GHCR 时，开发机用具备 `write:packages` 权限的令牌登录。

```bash
docker login ghcr.io
AUTOSTACK_PLATFORM=linux/amd64 ./release.sh v1.2.0
```

`release.sh` 从已提交的 Git 快照构建。新 Tag 对应当前提交，已有本地 Tag 对应它原来的提交，可用于为旧版本补发镜像。两个镜像都可用后，脚本才会把新 Tag 推到 `origin`。镜像的标签由完整 Git 提交和架构组成，例如 `<commit>-amd64`。

**版本号规范建议（语义化版本）：**

| 格式 | 说明 | 示例 |
|------|------|------|
| vX.Y.Z | 主版本.次版本.修订号 | v1.2.3 |
| 主版本 X | 不兼容的 API 变更 | v2.0.0 |
| 次版本 Y | 向下兼容的功能新增 | v1.3.0 |
| 修订号 Z | 向下兼容的问题修复 | v1.2.1 |

### 5.3 服务器更新流程

从源码构建流程迁移时，先用 `release.sh` 为线上当前 Git Tag 补发镜像。发布流程改动推送到 `origin/master` 后，在服务器的**原部署目录**更新部署文件。

```bash
git fetch --no-tags origin master
git restore --source=FETCH_HEAD -- .gitignore deploy.sh docker-compose.prod.yml
```

用 `prepare <当前 Tag>` 缓存回滚镜像。缺少可信的部署记录时，先用 `deploy <当前 Tag>` 部署并验收基线版本。首次应用新版 Compose 可能因 MySQL 健康检查变更而重建数据库容器，切换时沿用原数据目录并安排在迁移窗口内。

```bash
# 上线窗口之前执行，预先下载应用镜像
./deploy.sh prepare v1.2.0

# 到上线窗口执行，复用已下载的镜像
AUTOSTACK_OFFLINE=1 ./deploy.sh deploy v1.2.0
```

`AUTOSTACK_OFFLINE=1` 使用本地已有的 Git Tag 和镜像，适合 MySQL 已在运行的更新场景。也可以直接执行 `./deploy.sh deploy v1.2.0`，脚本会先补齐镜像，再更新容器。版本记录在健康检查通过后更新。单实例 Compose 切换时会重建应用容器，切换耗时主要取决于应用启动和数据库迁移。

### 5.4 版本回滚

可以用上一次成功部署的版本回滚。涉及数据库结构变更时，发版前备份数据并验证旧程序与新结构兼容。

```bash
# 回滚到上一个版本
./deploy.sh rollback

# 或指定回滚到某个版本
./deploy.sh deploy v1.1.0
```

## 六、常用运维命令

### 6.1 服务管理

```bash
# 启动已部署的版本
./deploy.sh start

# 停止服务
./deploy.sh stop

# 重启服务
./deploy.sh restart

# 查看状态
./deploy.sh status
```

### 6.2 日志查看

```bash
# 查看所有日志
./deploy.sh logs

# 查看后端日志
./deploy.sh logs backend

# 查看前端日志
./deploy.sh logs frontend

# 查看 MySQL 日志
docker compose -f docker-compose.prod.yml logs mysql
```

### 6.3 数据库操作

```bash
# 进入 MySQL 容器
docker exec -it autostack-mysql mysql -uroot -p

# 备份数据库
docker exec autostack-mysql mysqldump -uroot -p'你的密码' autostack > backup_$(date +%Y%m%d).sql

# 恢复数据库
docker exec -i autostack-mysql mysql -uroot -p'你的密码' autostack < backup.sql
```

### 6.4 数据备份

建议定期备份 MySQL 数据：

```bash
# 创建备份脚本
cat > /opt/AutoStack/backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/AutoStack/backups"
mkdir -p $BACKUP_DIR
docker exec autostack-mysql mysqldump -uroot -p"$MYSQL_PASSWORD" autostack | gzip > "$BACKUP_DIR/autostack_$(date +%Y%m%d_%H%M%S).sql.gz"
# 保留最近 7 天的备份
find $BACKUP_DIR -name "*.sql.gz" -mtime +7 -delete
EOF

chmod +x /opt/AutoStack/backup.sh

# 添加每日备份定时任务
echo "0 2 * * * root /opt/AutoStack/backup.sh" | sudo tee /etc/cron.d/autostack-backup
```

## 七、故障排查

### 7.1 服务无法启动

```bash
# 查看详细日志
docker compose -f docker-compose.prod.yml logs --tail=100

# 检查端口占用
sudo lsof -i :80
sudo lsof -i :3306

# 检查磁盘空间
df -h

# 检查内存
free -h
```

### 7.2 数据库连接失败

```bash
# 检查 MySQL 容器状态
docker ps -a | grep mysql

# 查看 MySQL 日志
docker logs autostack-mysql

# 测试数据库连接
docker exec -it autostack-mysql mysql -uroot -p -e "SELECT 1"
```

### 7.3 前端页面无法访问

```bash
# 检查前端容器
docker logs autostack-frontend

# 检查 Nginx 配置
docker exec autostack-frontend nginx -t

# 检查防火墙
sudo ufw status
sudo iptables -L -n
```

### 7.4 API 请求失败

```bash
# 检查后端日志
docker logs autostack-backend --tail=100

# 测试后端健康检查
docker exec autostack-backend wget -qO- http://127.0.0.1:8080/health

# 进入后端容器调试
docker exec -it autostack-backend sh
```

## 八、性能优化建议

1. **启用 Swap**（内存不足时）：
   ```bash
   sudo fallocate -l 2G /swapfile
   sudo chmod 600 /swapfile
   sudo mkswap /swapfile
   sudo swapon /swapfile
   echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
   ```

2. **调整 MySQL 配置**：根据服务器内存调整 `innodb_buffer_pool_size`

3. **使用 CDN**：将静态资源托管到 CDN 加速访问

4. **监控告警**：配置服务器监控（CPU、内存、磁盘）和服务健康检查

## 九、安全建议

1. **修改 SSH 默认端口**，禁用密码登录，使用密钥认证
2. **定期更新系统**：`sudo apt update && sudo apt upgrade`
3. **配置防火墙**，仅开放必要端口
4. **使用 HTTPS**，保护数据传输安全
5. **定期备份数据**，并验证备份可恢复性
6. **生产环境使用强密码**，不要使用默认密钥
