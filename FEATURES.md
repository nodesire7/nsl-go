# 功能实现清单

## ✅ 已完成功能

### 1. 用户系统
- [x] 用户注册（`/api/v1/auth/register`）
- [x] 用户登录（`/api/v1/auth/login`）
- [x] JWT Token认证
- [x] 用户信息查询（`/api/v1/profile`）
- [x] 密码加密存储（bcrypt）
- [x] 用户名和邮箱唯一性检查

### 2. 多域名支持
- [x] 用户自定义域名添加（`POST /api/v1/domains`）
- [x] 域名列表查询（`GET /api/v1/domains`）
- [x] 域名删除（`DELETE /api/v1/domains/:id`）
- [x] 设置默认域名（`PUT /api/v1/domains/:id/default`）
- [x] 创建链接时选择域名
- [x] 系统默认域名支持

### 3. 二维码生成
- [x] 创建链接时自动生成二维码
- [x] 二维码Base64格式返回
- [x] 支持256x256像素大小
- [x] 二维码存储在数据库中

### 4. 用户URL限制
- [x] 新用户默认限制10条链接
- [x] 创建链接前检查限制
- [x] 达到限制时返回友好错误提示
- [x] 管理员可提升用户限制（通过数据库）

### 5. 扩展功能
- [x] Redis缓存支持（可选）
- [x] API限流（每秒100个请求）
- [x] 分布式限流（使用Redis）
- [x] 本地限流（Redis不可用时）

### 6. 原有功能增强
- [x] 链接创建支持用户和域名
- [x] 链接列表按用户过滤
- [x] 链接删除权限检查
- [x] 哈希一致性检查按用户隔离

## 📝 API端点列表

### 认证相关
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/profile` - 获取用户信息

### 域名管理
- `POST /api/v1/domains` - 创建域名
- `GET /api/v1/domains` - 获取域名列表
- `DELETE /api/v1/domains/:id` - 删除域名
- `PUT /api/v1/domains/:id/default` - 设置默认域名

### 链接管理
- `POST /api/v1/links` - 创建短链接（自动生成二维码）
- `GET /api/v1/links` - 获取链接列表
- `GET /api/v1/links/search` - 搜索链接
- `GET /api/v1/links/:code` - 获取链接详情
- `DELETE /api/v1/links/:code` - 删除链接

### 统计和配置
- `GET /api/v1/stats` - 获取统计信息
- `GET /api/v1/settings` - 获取配置
- `PUT /api/v1/settings` - 更新配置

## 🔧 数据库变更

### 新增表
1. **users** - 用户表
   - id, username, email, password, role, max_links, created_at, updated_at

2. **domains** - 域名表
   - id, user_id, domain, is_default, is_active, created_at, updated_at

### 更新表
1. **links** - 链接表
   - 新增字段：user_id, domain_id, qr_code

## 🚀 部署说明

### 环境变量新增
- `REDIS_HOST` - Redis地址（可选）
- `REDIS_PASSWORD` - Redis密码（可选）

### Docker Compose更新
- 新增Redis服务
- 应用服务依赖Redis

## 📊 使用示例

### 1. 用户注册
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 2. 用户登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'
# 返回: {"token": "JWT_TOKEN", "user": {...}}
```

### 3. 添加域名
```bash
curl -X POST http://localhost:8080/api/v1/domains \
  -H "Authorization: Bearer JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "s.example.com",
    "is_default": true
  }'
```

### 4. 创建短链接（自动生成二维码）
```bash
curl -X POST http://localhost:8080/api/v1/links \
  -H "Authorization: Bearer JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://www.example.com",
    "title": "示例网站",
    "domain_id": 1
  }'
# 返回包含 qr_code 字段
```

## 🔐 权限说明

- **普通用户**：只能管理自己的链接和域名
- **管理员**：可以管理所有链接（通过API Token或admin角色）
- **新用户限制**：默认最多10条链接，达到限制后需要联系管理员提升

## 📝 注意事项

1. JWT Token有效期为24小时
2. 二维码生成失败不影响链接创建
3. Redis是可选的，不配置也能正常运行
4. 域名验证需要用户自行配置DNS
5. 系统默认域名使用BASE_URL配置

