# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
# - git：部分依赖可能需要走 VCS 拉取
# - ca-certificates：Alpine 默认可能缺失 CA，导致 go mod download TLS 失败
RUN apk add --no-cache git ca-certificates && update-ca-certificates

# 复制 go.mod（go.sum 可能不存在：允许在容器内通过 tidy 生成）
COPY go.mod ./

# 先下载依赖（加速缓存命中；如无 go.sum 也可正常运行）
RUN go mod download

# 复制源代码
COPY . .

# 确保依赖与校验和完整（容器内生成/更新 go.sum）
RUN go mod tidy

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o nsl-go ./cmd/api

# 运行阶段
FROM alpine:latest

# 安装ca证书
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/nsl-go .

# 复制静态文件
COPY --from=builder /app/web ./web

# 暴露端口
EXPOSE 9110

# 运行应用
CMD ["./nsl-go"]

