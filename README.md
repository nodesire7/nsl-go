# 🔗 New short link (NSL GO)

一个功能完整的短链接生成和管理系统，使用Go语言重构，支持PostgreSQL数据库和Meilisearch全文搜索。

## ✨ 特性

* 🚀 **高性能**: Go语言编写，性能优异
* 🗄️ **PostgreSQL**: 使用PostgreSQL作为主数据库
* 🔍 **全文搜索**: 集成Meilisearch，支持快速搜索
* 🔢 **动态链接长度**: 自动扩展链接长度（6位起，用完自动扩展）
* 🔐 **内容哈希一致性**: 同一用户同一域名下，相同URL返回相同短链接（幂等粒度：`user + domain + hash`）
* 📊 **数据统计**: 完整的访问统计和分析
* 🎨 **Web UI**: 美观的前台管理面板，支持登录页面
* 🐳 **Docker部署**: 一键部署，零配置
* 👤 **用户系统**: 支持用户注册、登录、JWT认证
* 🔑 **用户Token**: 每个用户自动生成永久Token，用于API调用
* 👨‍💼 **Admin管理**: 自动创建admin用户，提供命令行管理工具
* 🌐 **多域名支持**: 每个用户可自定义多个短链接域名
* 📱 **二维码生成**: 自动生成短链接二维码
* ⚡ **Redis缓存**: 支持Redis缓存提升性能
* 🛡️ **API限流**: 内置限流保护，防止滥用
* 🔒 **权限控制**: 新用户默认限制10条链接，可联系管理员提升

## 🚀 快速开始

### 一键安装（推荐）

```bash
curl -fsSL https://raw.githubusercontent.com/nodesire7/nsl-go/main/install.sh | bash
```

### 使用Docker Compose

```bash
docker-compose up -d
```

**首次启动后**：
1. 查看日志确认已创建 admin 用户：
   ```bash
   docker-compose logs app | grep "Admin用户已创建"
   ```
2. 出于安全原因，**不会在日志中打印明文密码/Token**。请使用管理工具生成/重置 admin 密码后登录：
   ```bash
   make build-admin
   ./bin/nsl-admin -action=reset-password
   ```
3. 访问 `http://localhost:9110/login` 登录

### 手动安装

1. 从 [Releases](https://github.com/nodesire7/nsl-go/releases) 下载对应平台的二进制文件
2. 解压并运行：

```bash
tar -xzf nsl-go-linux-amd64.tar.gz
./nsl-go
```

**首次启动后**：
- 查看控制台输出确认已创建 admin 用户
- 使用管理工具重置 admin 密码后登录：
  ```bash
  ./bin/nsl-admin -action=reset-password
  ```
- 访问 `http://localhost:9110/login` 登录

### Docker Hub

```bash
docker pull nodesire77/nsl-go:latest
```

## 📋 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `BASE_URL` | http://localhost:9110 | 服务基础URL |
| `JWT_SECRET` | 必需 | **Cookie 登录鉴权**的JWT签名密钥（建议 `openssl rand -hex 32`） |
| `DB_HOST` | localhost | PostgreSQL主机 |
| `DB_PORT` | 5432 | PostgreSQL端口 |
| `DB_USER` | postgres | 数据库用户 |
| `DB_PASSWORD` | postgres | 数据库密码 |
| `DB_NAME` | shortlink | 数据库名 |
| `MEILI_HOST` | http://localhost:7700 | Meilisearch地址 |
| `MEILI_KEY` | | Meilisearch主密钥 |
| `REDIS_HOST` | | Redis地址（可选） |
| `REDIS_PASSWORD` | | Redis密码（可选） |
| `MIN_CODE_LENGTH` | 6 | 最小短代码长度 |
| `MAX_CODE_LENGTH` | 10 | 最大短代码长度 |
| `LOG_LEVEL` | INFO | 日志级别 |
| `SERVER_PORT` | 9110 | 服务端口 |

## ⚠️ 重要说明（请务必读）

### 多域名重定向（按 Host 解析）

- **当你使用自定义短链域名时**，服务端会根据请求的 `Host`（访问的域名）去匹配 `domains.domain`，然后再用 `(domain_id, code)` 精确查询，避免多域名下 code 冲突导致误跳转。
- 如果请求 Host 无法匹配任何 domain：会回退到“全库按 code 查询”，**仅当全库只命中 1 条**才允许跳转，否则返回 404。

> 建议：`domains.domain` 保存为纯域名（例如 `s.example.com`），不要带路径；如果是本地测试带端口，也支持 `localhost:9110` 的匹配。

### 依赖校验（go.sum）

当前仓库可能尚未提交 `go.sum`。CI 已做兼容处理，但**建议你在本地安装 Go 后补齐并提交**：

```bash
go mod tidy
git add go.sum
git commit -m "chore: add go.sum"
git push
```

## 🔧 API接口

### 认证方式

系统支持两种认证方式：

1. **用户API Token**（推荐，永久有效）：
```
Authorization: Bearer nsl_xxxxxxxxxxxxx
```

2. **JWT Token**（用于Web登录，HttpOnly Cookie）：
```
Authorization: Bearer YOUR_JWT_TOKEN
```
或通过 Cookie：`Cookie: access_token=YOUR_JWT_TOKEN`

### 用户注册

```bash
curl -X POST http://localhost:9110/api/v2/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

**响应**（包含用户的API Token，永久有效，仅返回一次）：
```json
{
  "token": "JWT_TOKEN",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "api_token": "nsl_xxxxxxxxxxxxx",
    "role": "user",
    "max_links": 10
  }
}
```

> ⚠️ **重要**：`api_token` 仅在注册时返回一次，请妥善保存。后续登录/资料接口不会返回 `api_token`（已改为 hash 存储）。

### 用户登录
> 注意：登录接口不再返回长期 `api_token`。如需创建/轮换 API Token，请调用 `/api/v2/profile/token`。

```bash
curl -X POST http://localhost:9110/api/v2/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

**响应**（Web UI 会设置 HttpOnly Cookie）：
```json
{
  "token": "JWT_TOKEN",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "role": "user",
    "max_links": 10
  }
}
```

### 更新用户Token

```bash
curl -X POST http://localhost:9110/api/v2/profile/token \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "X-CSRF-Token: YOUR_CSRF_TOKEN"
```

### 创建短链接

```bash
curl -X POST http://localhost:9110/api/v2/links \
  -H "Authorization: Bearer nsl_xxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.example.com",
    "title": "示例网站",
    "code": "custom",
    "domain_id": 1
  }'
```

**响应包含二维码**：
```json
{
  "id": 1,
  "code": "custom",
  "short_url": "https://s.example.com/custom",
  "original_url": "https://www.example.com",
  "title": "示例网站",
  "qr_code": "data:image/png;base64,iVBORw0KGgo...",
  "click_count": 0,
  "created_at": "2025-01-XX..."
}
```

### 获取链接列表

```bash
curl -X GET "http://localhost:9110/api/v2/links?page=1&limit=20" \
  -H "Authorization: Bearer nsl_xxxxxxxxxxxxx"
```

### 搜索链接

```bash
curl -X GET "http://localhost:9110/api/v2/links/search?q=example" \
  -H "Authorization: Bearer nsl_xxxxxxxxxxxxx"
```

### 删除链接

```bash
curl -X DELETE "http://localhost:9110/api/v2/links/custom" \
  -H "Authorization: Bearer nsl_xxxxxxxxxxxxx" \
  -H "X-CSRF-Token: YOUR_CSRF_TOKEN"
```

### 获取统计信息

```bash
curl -X GET "http://localhost:9110/api/v2/stats" \
  -H "Authorization: Bearer nsl_xxxxxxxxxxxxx"
```

## ✅ redo.md 完成度对照（当前仓库状态）

- **已完成**
  - ✅ `JWT_SECRET` 强制配置；移除“系统级 API_TOKEN 超级通行证”
  - ✅ Web UI：HttpOnly Cookie + SameSite + CSRF（双提交）
  - ✅ 短码生成：`crypto/rand` + DB 唯一约束冲突重试（并发安全）
  - ✅ 幂等：按 `(user_id, domain_id, hash)` 粒度返回已有短链
  - ✅ Redis：热点重定向缓存（v2 已按域名隔离缓存 key）
  - ✅ 安全头、基础 SSRF 校验、请求 request_id、限流中间件
  - ✅ 重写架构：`internal/config + internal/db(pgxpool) + internal/repo + internal/service + internal/httpv2`
  - ✅ **统计写入异步化**：使用 `internal/jobs` worker 批量写入点击数/访问日志，跳转路径极速化
  - ✅ **API Token 存储**：已停止写入 `users.api_token` 明文字段，历史数据会回填 `api_token_hash` 并清空明文列；鉴权优先按 hash 匹配
  - ✅ **V1 代码完全删除**：已彻底删除 legacy 代码，全面迁移到 `internal/*` 架构

- **已完成（全部）**
  - ✅ **审计日志**：管理员操作、敏感操作记录（redo.md 1.2, 2.7, 7.1）
  - ✅ **RBAC 权限点**：细粒度权限点（`link:create`, `link:delete`, `link:view`, `link:list`, `stats:view` 等）（redo.md 4.2, 6.2）
  - ✅ **Meilisearch 写入失败补偿/重试/后台任务**：异步队列 + 重试机制（最大3次，间隔5秒）（redo.md 2.6）
  - ✅ **结构化日志统一**：已统一使用 `utils` logger，移除所有 `log.Printf`（redo.md 2.7）
  - ✅ **集成测试**：使用 testcontainers 实现 PG/Redis 集成测试（redo.md 6.3）
  - ✅ **CI 质量工具**：`golangci-lint` / `gosec` 已在 CI 中落地（redo.md 6.3）
  - ✅ **Metrics（Prometheus）**：指标收集和暴露（HTTP 请求、业务指标、限流等）（redo.md 6.3）
  - ✅ **Tracing**：分布式追踪（OpenTelemetry + Jaeger）（redo.md 3.1）
  - ✅ **聚合统计扩展**：日/周/月、来源、UA、IP 等维度统计（redo.md 5.3）
  - ✅ **限流策略优化**：滑动窗口 + 令牌桶算法（redo.md 2.5）
  - ✅ **代理链路真实 IP 处理**：正确处理 X-Forwarded-For / X-Real-IP（redo.md 2.5）

## 🔑 用户Token说明

- **自动生成**: 用户注册时自动生成永久Token（格式：`nsl_xxxxxxxxxxxxx`）
- **永久有效**: Token没有过期时间，除非：
  - 用户被删除
  - 用户主动更新Token（通过 `/api/v2/profile/token` 接口）
- **用途**: 用于API调用，替代JWT Token进行长期访问
- **安全**: Token 以 SHA256 hash 存储在数据库中，不再存储明文；建议定期更新

## 👤 Admin用户管理

### 自动创建Admin用户

系统首次启动时会**自动创建admin用户**。出于安全原因，日志中**不输出明文密码/Token**，请使用管理工具重置密码：

```
✅ Admin用户已创建（出于安全原因，不在日志中输出明文密码/Token）
```

### 重置Admin密码

使用管理工具重置admin密码：

```bash
# 编译管理工具
make build-admin
# 或
go build -o bin/nsl-admin ./cmd/admin

# 随机生成新密码（推荐）
./bin/nsl-admin -action=reset-password

# 指定新密码
./bin/nsl-admin -action=reset-password -password=MyNewPassword123

# 查看admin用户信息
./bin/nsl-admin -action=show-info
```

**Windows用户**：
```powershell
# 编译
go build -o bin\nsl-admin.exe ./cmd/admin

# 使用
.\bin\nsl-admin.exe -action=reset-password
.\bin\nsl-admin.exe -action=show-info
```

### 登录页面

访问 `http://localhost:9110/login` 进入登录页面，使用admin账户登录。

**首次登录后建议**：
1. 修改admin密码（使用管理工具）
2. 创建普通用户账户
3. 妥善保管API Token

## 🎨 Web UI

### 登录

访问 `http://localhost:9110/login` 进入登录页面。

**默认admin账户**：
- 用户名：`admin`
- 密码：请使用管理工具重置生成（不会写入日志）

### 管理面板

登录后访问 `http://localhost:9110` 查看Web管理面板，可以：
- 📊 查看统计信息
- 🔗 创建和管理短链接
- 🔍 搜索链接
- 📱 查看二维码
- ⚙️ 管理域名设置

## 📄 许可证

MIT License
