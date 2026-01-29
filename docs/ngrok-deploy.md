# ngrok 内网穿透部署指南

本文档记录如何使用 ngrok 将本地 AutoStack 服务暴露到公网访问。

## 前置条件

- Docker 和 Docker Compose 已安装
- ngrok 已安装（`brew install ngrok`）
- ngrok 账号已注册（https://ngrok.com）

## 一、安装 ngrok

```bash
# macOS
brew install ngrok

# 如果 brew 下载失败，设置代理后重试
export https_proxy=http://127.0.0.1:7890
export http_proxy=http://127.0.0.1:7890
brew install ngrok
```

## 二、配置 ngrok authtoken

1. 登录 https://dashboard.ngrok.com/get-started/your-authtoken
2. 复制你的 authtoken
3. 配置：

```bash
ngrok config add-authtoken <your-authtoken>
```

## 三、启动服务

### 3.1 启动 Docker 服务

```bash
cd /Applications/workspace/AutoStack

# 首次启动（构建镜像）
docker-compose up -d --build

# 后续启动
docker-compose up -d
```

### 3.2 验证服务状态

```bash
# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

## 四、启动 ngrok 穿透

### 4.1 前台运行（测试用）

```bash
ngrok http 80
```

启动后显示公网地址：
```
Forwarding    https://xxxx-xx-xx.ngrok-free.app -> http://localhost:80
```

### 4.2 后台运行（长期使用）

**方法一：nohup**

```bash
nohup ngrok http 80 > /dev/null 2>&1 &

# 查看公网地址：访问 http://localhost:4040
```

**方法二：screen（推荐）**

```bash
# 创建会话
screen -S ngrok

# 运行 ngrok
ngrok http 80

# 分离会话：按 Ctrl+A 然后按 D

# 重新连接
screen -r ngrok

# 关闭会话
screen -X -S ngrok quit
```

**方法三：tmux**

```bash
# 创建会话
tmux new -s ngrok

# 运行 ngrok
ngrok http 80

# 分离会话：按 Ctrl+B 然后按 D

# 重新连接
tmux attach -t ngrok

# 关闭会话
tmux kill-session -t ngrok
```

## 五、常用命令

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 查看指定服务日志
docker-compose logs -f backend
docker-compose logs -f frontend

# 重新构建并启动
docker-compose up -d --build
```

## 六、注意事项

1. **ngrok 免费版限制**：
   - 每次重启地址会变化
   - 有连接数和带宽限制
   - 首次访问会显示 ngrok 警告页面

2. **保持服务运行**：
   - Docker 服务使用 `-d` 参数已在后台运行
   - ngrok 需要使用 screen/tmux/nohup 保持后台运行

3. **端口说明**：
   - 80：前端 nginx（对外唯一入口）
   - 8080：后端 API（仅 Docker 内部访问）
   - 3306：MySQL（仅本地访问）
   - 4040：ngrok Web 界面（本地查看状态）

## 七、故障排查

### 端口被占用

```bash
# 查找占用端口的进程
lsof -i :80
lsof -i :8080

# 结束进程
kill -9 <PID>
```

### ngrok authtoken 无效

重新从 https://dashboard.ngrok.com/get-started/your-authtoken 获取并配置。

### 服务无法访问

```bash
# 检查容器状态
docker-compose ps

# 检查容器日志
docker-compose logs backend
docker-compose logs frontend

# 重启服务
docker-compose restart
```
