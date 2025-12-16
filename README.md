# 🔗 New short link (NSL GO)

一个功能完整的短链接生成和管理系统，使用Go语言重构，支持PostgreSQL数据库和Meilisearch全文搜索。

## ✨ 特性

* 🚀 **高性能**: Go语言编写，性能优异
* 🗄️ **PostgreSQL**: 使用PostgreSQL作为主数据库
* 🔍 **全文搜索**: 集成Meilisearch，支持快速搜索
* 🔢 **动态链接长度**: 自动扩展链接长度（6位起，用完自动扩展）
* 🔐 **内容哈希一致性**: 相同URL返回相同短链接
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
1. 查看日志获取admin用户密码：
   ```bash
   docker-compose logs app | grep "Admin用户已创建"
   ```
2. 访问 `http://localhost:9110/login` 登录

### 手动安装

1. 从 [Releases](https://github.com/nodesire7/nsl-go/releases) 下载对应平台的二进制文件
2. 解压并运行：

```bash
tar -xzf nsl-go-linux-amd64.tar.gz
./nsl-go
```

**首次启动后**：
- 查看控制台输出，获取admin用户密码
- 访问 `http://localhost:9110/login` 登录

### Docker Hub

```bash
docker pull nodesire7/nsl-go:latest
```

## 📋 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `BASE_URL` | http://localhost:9110 | 服务基础URL |
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

系统支持三种认证方式：

1. **用户API Token**（推荐，永久有效）：
```
Authorization: Bearer nsl_xxxxxxxxxxxxx
```

2. **JWT Token**（用于Web登录）：
```
Authorization: Bearer YOUR_JWT_TOKEN
```

3. **系统API Token**（管理员）：
```
Authorization: Bearer YOUR_API_TOKEN
```

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
- **安全**: Token存储在数据库中，建议定期更新

## 👤 Admin用户管理

### 自动创建Admin用户

系统首次启动时会**自动创建admin用户**，密码会在日志中输出：

```
==========================================
✅ Admin用户已创建
==========================================
用户名: admin
密码: [随机生成的16位密码]
API Token: nsl_xxxxxxxxxxxxx
==========================================
⚠️  请妥善保管以上信息，建议首次登录后修改密码
==========================================
```

### 重置Admin密码

使用管理工具重置admin密码：

```bash
# 编译管理工具
make build-admin
# 或
go build -o nsl-admin ./cmd/admin

# 随机生成新密码（推荐）
./nsl-admin -action=reset-password

# 指定新密码
./nsl-admin -action=reset-password -password=MyNewPassword123

# 查看admin用户信息
./nsl-admin -action=show-info
```

**Windows用户**：
```powershell
# 编译
go build -o nsl-admin.exe ./cmd/admin

# 使用
.\nsl-admin.exe -action=reset-password
.\nsl-admin.exe -action=show-info
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
- 密码：首次启动时在日志中显示（随机生成）

### 管理面板

登录后访问 `http://localhost:9110` 查看Web管理面板，可以：
- 📊 查看统计信息
- 🔗 创建和管理短链接
- 🔍 搜索链接
- 📱 查看二维码
- ⚙️ 管理域名设置

## 📄 许可证

MIT License
