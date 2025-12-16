# 🚀 部署和配置指南

## GitHub Actions Secrets 配置

为了让CI/CD正常工作，需要在GitHub仓库中配置以下Secrets：

### 1. 进入仓库设置
访问：https://github.com/nodesire7/nsl-go/settings/secrets/actions

### 2. 添加以下Secrets

#### DOCKERHUB_USERNAME
- **名称**: `DOCKERHUB_USERNAME`
- **值**: 你的Docker Hub用户名（例如：`nodesire7`）

#### DOCKERHUB_TOKEN
- **名称**: `DOCKERHUB_TOKEN`
- **值**: Docker Hub访问令牌

**获取Docker Hub Token步骤**：
1. 登录 Docker Hub: https://hub.docker.com/
2. 点击右上角头像 → Account Settings（账户设置）
3. 左侧菜单选择 Security（安全）
4. 点击 "New Access Token"（新建访问令牌）
5. 输入描述（如：GitHub Actions）
6. 选择权限：Read & Write（读写）
7. 复制生成的Token

### 3. 验证配置

配置完成后，每次推送到main分支都会：
- ✅ 自动构建Docker镜像并推送到Docker Hub
- ✅ 自动构建多平台二进制文件（Linux、Windows、macOS）
- ✅ 自动创建GitHub Releases（发布版本）

## 本地开发环境

### 环境变量配置

创建 `.env` 文件：

```env
# API配置
API_TOKEN=your-secret-api-token-here
BASE_URL=http://localhost:9110

# PostgreSQL配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=shortlink
DB_SSLMODE=disable

# Meilisearch配置
MEILI_HOST=http://localhost:7700
MEILI_KEY=

# Redis配置（可选）
REDIS_HOST=localhost:6379
REDIS_PASSWORD=

# 短链接配置
MIN_CODE_LENGTH=6
MAX_CODE_LENGTH=10

# 日志配置
LOG_LEVEL=INFO

# 服务器配置
SERVER_PORT=9110
SERVER_MODE=release
```

### 启动服务

```bash
# 使用Docker Compose（推荐）
docker-compose up -d

# 或手动启动
go run cmd/server/main.go
```

## 生产环境部署

### Docker部署

```bash
# 拉取最新镜像
docker pull nodesire7/nsl-go:latest

# 运行容器
docker run -d \
  --name nsl-go \
  -p 9110:9110 \
  -e API_TOKEN=your-token \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=password \
  nodesire7/nsl-go:latest
```

### 二进制文件部署

从 [Releases](https://github.com/nodesire7/nsl-go/releases) 下载对应平台的二进制文件：

```bash
# Linux
wget https://github.com/nodesire7/nsl-go/releases/download/latest/nsl-go-linux-amd64.tar.gz
tar -xzf nsl-go-linux-amd64.tar.gz
./nsl-go
```

## 验证部署

访问健康检查端点：

```bash
curl http://localhost:9110/health
```

应该返回：
```json
{
  "status": "ok",
  "service": "short-link"
}
```

## 下一步

1. ✅ 配置GitHub Secrets（DOCKERHUB_USERNAME和DOCKERHUB_TOKEN）
2. ✅ 测试API功能
3. ✅ 配置域名和DNS
4. ✅ 设置SSL证书（生产环境）

