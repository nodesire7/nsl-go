# 🔗 New short link (NSL GO)

一个功能完整的短链接生成和管理系统，使用Go语言重构，支持PostgreSQL数据库和Meilisearch全文搜索。

## ✨ 特性

* 🚀 **高性能**: Go语言编写，性能优异
* 🗄️ **PostgreSQL**: 使用PostgreSQL作为主数据库
* 🔍 **全文搜索**: 集成Meilisearch，支持快速搜索
* 🔢 **动态链接长度**: 自动扩展链接长度（6位起，用完自动扩展）
* 🔐 **内容哈希一致性**: 同一用户在同一域名下提交相同URL，返回已存在短链接（重写版幂等粒度：`user_id + domain_id + hash`）
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
| `LOG_LEVEL` | INFO | 日志级别 |
| `SERVER_PORT` | 9110 | 服务端口 |

## 🔧 API接口

### 认证方式

系统支持两种认证方式：

1. **用户API Token**（推荐，永久有效）：
```
Authorization: Bearer nsl_xxxxxxxxxxxxx
```

2. **JWT Token**（用于Web登录）：
```
Authorization: Bearer YOUR_JWT_TOKEN
```

> 说明：旧版曾支持 `API_TOKEN` 作为“系统通行证”，存在高风险（泄漏即全站失守），重写版将移除该设计。

### v1 / v2 说明（重写增量迁移）

- **`/api/v1`**：旧实现（legacy），功能齐全但分层/安全基线仍在逐步迁移中。
- **`/api/v2`**：重写版（internal/* 分层 + pgxpool），优先迁移核心链路：
  - `POST /api/v2/auth/register` `POST /api/v2/auth/login` `POST /api/v2/auth/logout`
  - `GET /api/v2/profile` `POST /api/v2/profile/token`
  - `POST /api/v2/links` `GET /api/v2/links`
- **跳转 `/:code`**：已优先走重写版 v2 的解析逻辑（按 `Host` 解析 domain，避免多域名下同 code 误跳转），并写入点击数与访问日志，支持 Redis 热点缓存。

### 用户注册

```bash
curl -X POST http://localhost:9110/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

**响应包含用户的API Token**（永久有效）：
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

### 用户登录
> 注意：登录接口不再返回长期 `api_token`。如需创建/轮换 API Token，请调用 `/api/v1/profile/token`。

```bash
curl -X POST http://localhost:9110/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
```

### 更新用户Token

```bash
curl -X POST http://localhost:9110/api/v1/profile/token \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 创建域名

```bash
curl -X POST http://localhost:9110/api/v1/domains \
  -H "Authorization: Bearer nsl_xxxxxxxxxxxxx" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "s.example.com",
    "is_default": true
  }'
```

### 创建短链接

```bash
curl -X POST http://localhost:9110/api/v1/links \
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
  "success": true,
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
curl -X GET "http://localhost:9110/api/v1/links?page=1&limit=20" \
  -H "Authorization: Bearer nsl_xxxxxxxxxxxxx"
```

### 搜索链接

```bash
curl -X GET "http://localhost:9110/api/v1/links/search?q=example" \
  -H "Authorization: Bearer nsl_xxxxxxxxxxxxx"
```

## 🔑 用户Token说明

- **自动生成**: 用户注册时自动生成永久Token（格式：`nsl_xxxxxxxxxxxxx`）
- **永久有效**: Token没有过期时间，除非：
  - 用户被删除
  - 用户主动更新Token（通过 `/api/v1/profile/token` 接口）
- **用途**: 用于API调用，替代JWT Token进行长期访问
- **安全**: 当前实现为兼容迁移阶段，会同时写入 `api_token_hash`（用于匹配）与 `api_token`（旧字段）。`redo.md` 目标是最终仅保留 **hash 存储**（避免数据库泄漏导致 token 直接可用），后续会继续迁移。

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

---

## ✅ `redo.md` 对照清单（当前完成度）

> 说明：本项目正在做“增量重写”，因此会同时存在 v1（legacy）与 v2（重写版）实现。

### 已完成

- **安全基线**
  - **JWT_SECRET 必须配置**（未设置直接启动失败）
  - **移除系统级 API_TOKEN 超级通行证**
  - **Web UI：HttpOnly Cookie + CSRF（双提交 Cookie）**
  - **基础安全头**（`SecurityHeadersMiddleware`）
  - **URL SSRF 基础校验**（仅允许 http/https + 内网拦截）
- **并发正确性**
  - **短码生成使用 crypto/rand**（拒绝采样）
  - **DB 唯一约束冲突重试**（并发安全）
- **性能**
  - **跳转路径 Redis 热点缓存**（code -> url + link_id）
- **架构**
  - 已引入 `internal/config` `internal/db(pgxpool)` `internal/repo` `internal/service` `internal/httpv2`
  - `/api/v2` 已迁移：用户鉴权、短链创建/列表、跳转
- **可观测（部分）**
  - `request_id` 中间件已加入

### 进行中 / 未完成（后续计划）

- **Token 存储完全去明文**：目前仍保留 `api_token` 明文字段用于兼容；目标是最终仅存 `api_token_hash`
- **RBAC 权限点**：当前仍以 `admin/user` 角色为主，权限点体系待补齐
- **审计日志**：管理员/敏感操作的审计日志待实现
- **异步统计/worker**：当前跳转写入为同步 best-effort；`redo.md` 建议改为队列/worker/batch 聚合
- **Meilisearch 一致性补偿**：写入失败重试/补偿/死信队列待实现
- **质量门禁**：CI 目前跑 `go test ./...`，但 `golangci-lint` / `gosec` / Prometheus metrics 尚未接入
