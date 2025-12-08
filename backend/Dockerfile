# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 安装依赖
RUN apk add --no-cache gcc musl-dev

# 复制 go.mod
COPY go.mod ./

# 复制源代码
COPY . .

# 下载依赖并构建
RUN go mod tidy && CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o autostack ./cmd/server

# 运行阶段
FROM alpine:3.19

WORKDIR /app

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata wget

# 设置时区
ENV TZ=Asia/Shanghai

# 创建数据目录
RUN mkdir -p /app/data

# 从构建阶段复制二进制文件
COPY --from=builder /build/autostack .
COPY --from=builder /build/config.yaml .

# 暴露端口
EXPOSE 8080

# 启动服务
CMD ["./autostack"]

