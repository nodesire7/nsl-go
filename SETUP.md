# 🚀 部署和配置指南

## GitHub Actions Secrets 配置

为了让CI/CD正常工作，需要在GitHub仓库中配置以下Secrets：

### 1. 进入仓库设置
访问：https://github.com/nodesire7/nsl-go/settings/secrets/actions

### 2. 添加以下Secrets

#### DOCKERHUB_USERNAME（可选）
- **名称**: `DOCKERHUB_USERNAME`
- **值**: 你的Docker Hub用户名（例如：`nodesire77`）
- **说明**: 如果不需要推送Docker镜像到Docker Hub，可以跳过此配置

#### DOCKERHUB_TOKEN（可选）
- **名称**: `DOCKERHUB_TOKEN`
- **值**: Docker Hub访问令牌
- **说明**: 如果不需要推送Docker镜像到Docker Hub，可以跳过此配置

#### GITHUB_TOKEN（通常不需要手动配置）
- **说明**: GitHub Actions 会自动提供 `GITHUB_TOKEN`，通常不需要手动配置
- **如果遇到 403 权限错误**，请检查仓库设置：
  1. 访问：https://github.com/nodesire7/nsl-go/settings/actions
  2. 找到 "Workflow permissions"（工作流权限）部分
  3. 确保设置为：
     - ✅ "Read and write permissions"（读写权限）
     - ✅ "Allow GitHub Actions to create and approve pull requests"（允许创建和批准PR）
  4. 如果仍然失败，可以创建 Personal Access Token (PAT)：
     - 访问：https://github.com/settings/tokens
     - 点击 "Generate new token" → "Generate new token (classic)"
     - 选择权限：`repo`（完整仓库访问权限）
     - 复制生成的 token
     - 在 GitHub Secrets 中添加：`GITHUB_TOKEN` = 你的 PAT

**重要：必须先创建Docker Hub仓库！**

在配置Token之前，请先创建Docker Hub仓库：

1. 登录 Docker Hub: https://hub.docker.com/
2. 点击右上角 "+" → "Create Repository"（创建仓库）
3. 仓库名称填写：`nsl-go`（完整路径为：`nodesire77/nsl-go`）
4. 选择可见性：Public（公开）或 Private（私有）
5. 点击 "Create"（创建）

**获取Docker Hub Token步骤**：
1. 登录 Docker Hub: https://hub.docker.com/
2. 点击右上角头像 → Account Settings（账户设置）
3. 左侧菜单选择 Security（安全）
4. 点击 "New Access Token"（新建访问令牌）
5. 输入描述（如：GitHub Actions）
6. **选择权限：Read, Write & Delete**（读写删除，必须包含推送权限）
   - ⚠️ 如果只选择 Read，将无法推送镜像
   - ⚠️ 必须至少包含 Write 权限
7. 点击 "Generate"（生成）
8. 复制生成的Token（只显示一次，请妥善保存）
9. 将Token粘贴到GitHub Secrets的 `DOCKERHUB_TOKEN` 中

**常见问题排查**：

如果遇到 "push access denied" 或 "repository does not exist" 或 "insufficient_scope" 错误：

### 步骤 1: 确认仓库已创建 ⭐ 最重要

1. 访问：https://hub.docker.com/r/nodesire77/nsl-go
2. 如果显示 **404 Not Found**，说明仓库不存在，需要先创建：
   - 访问：https://hub.docker.com/
   - 点击右上角 "+" → "Create Repository"
   - 仓库名称：`nsl-go`
   - 可见性：Public 或 Private
   - 点击 "Create"
3. 创建后，再次访问 https://hub.docker.com/r/nodesire77/nsl-go 应该能看到仓库页面

### 步骤 2: 确认Token权限 ⭐ 必须包含 Write

错误信息 `insufficient_scope` 通常表示 Token 权限不足。

1. 登录 Docker Hub: https://hub.docker.com/
2. 点击右上角头像 → Account Settings → Security
3. 找到你的 Token，检查权限：
   - ❌ 如果只有 "Read" 权限 → 无法推送
   - ✅ 必须有 "Write" 或 "Read, Write & Delete" 权限
4. 如果权限不足：
   - 删除旧 Token
   - 创建新 Token，**必须选择 "Read, Write & Delete"**
   - 复制新 Token
   - 更新 GitHub Secrets 中的 `DOCKERHUB_TOKEN`

### 步骤 3: 验证Token有效性

1. 检查 Token 是否过期
2. 确认 Token 格式正确（应该以 `dckr_pat_` 开头）
3. 如果 Token 过期，重新生成并更新 GitHub Secrets

### 步骤 4: 确认Secrets配置

在 GitHub 仓库设置中检查：
- `DOCKERHUB_USERNAME` = `nodesire7`（你的 Docker Hub 用户名，不含 `@` 符号）
- `DOCKERHUB_TOKEN` = 完整的 Token 字符串（以 `dckr_pat_` 开头）

### 快速检查清单

- [ ] Docker Hub 仓库已创建（访问 https://hub.docker.com/r/nodesire77/nsl-go 能看到页面）
- [ ] Token 权限包含 "Write" 或 "Read, Write & Delete"
- [ ] Token 未过期
- [ ] GitHub Secrets 中 `DOCKERHUB_USERNAME` 和 `DOCKERHUB_TOKEN` 都已正确配置
- [ ] `DOCKERHUB_USERNAME` 不包含 `@` 符号（只是用户名，不是邮箱）

**如果以上都确认无误，但仍有问题，请检查：**
- Token 是否被意外删除或撤销
- Docker Hub 账户是否被限制
- 网络连接是否正常

### 3. 验证配置

配置完成后，每次推送到main分支都会：
- ✅ 自动构建Docker镜像并推送到Docker Hub
- ✅ 自动构建多平台二进制文件（Linux、Windows、macOS）
- ✅ 自动创建GitHub Releases（发布版本）

## 本地开发环境

### 环境变量配置

创建 `.env` 文件：

```env
# 鉴权配置（必需）
# Cookie 登录鉴权的 JWT 签名密钥（建议 openssl rand -hex 32）
JWT_SECRET=your-jwt-secret-here
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
go run cmd/api/main.go
```

## 生产环境部署

### Docker部署

```bash
# 拉取最新镜像
docker pull nodesire77/nsl-go:latest

# 运行容器
docker run -d \
  --name nsl-go \
  -p 9110:9110 \
  -e JWT_SECRET=your-jwt-secret \
  -e DB_HOST=postgres \
  -e DB_PASSWORD=password \
  nodesire77/nsl-go:latest
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

