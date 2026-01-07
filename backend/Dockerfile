# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 设置Alpine国内镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 设置Go代理为国内镜像
ENV GOPROXY=https://goproxy.cn,direct

# 复制 go.mod
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建（使用MySQL时不需要CGO）
RUN CGO_ENABLED=0 GOOS=linux go build -o autostack ./cmd/server

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 设置Alpine国内镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget

# 设置时区为上海
ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建数据目录
RUN mkdir -p /app/data

# 从构建阶段复制二进制文件
COPY --from=builder /build/autostack .
COPY --from=builder /build/config.yaml .

# 暴露端口
EXPOSE 8080

# 启动服务
CMD ["./autostack"]
