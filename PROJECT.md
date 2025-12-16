# 项目结构说明

## 📁 目录结构

```
short-link/
├── cmd/
│   └── server/
│       └── main.go              # 主程序入口
├── config/
│   └── config.go                 # 配置管理
├── database/
│   ├── database.go               # 数据库连接和迁移
│   ├── link_repository.go       # 链接数据访问层
│   ├── access_log_repository.go  # 访问日志数据访问层
│   └── settings_repository.go    # 配置数据访问层
├── models/
│   ├── link.go                   # 链接模型
│   └── access_log.go             # 访问日志模型
├── services/
│   ├── link_service.go           # 链接业务逻辑
│   └── search_service.go         # 搜索服务（Meilisearch）
├── handlers/
│   ├── link_handler.go           # 链接HTTP处理器
│   ├── stats_handler.go          # 统计HTTP处理器
│   └── settings_handler.go        # 配置HTTP处理器
├── middleware/
│   ├── auth.go                   # 认证中间件
│   └── logger.go                 # 日志中间件
├── utils/
│   └── logger.go                 # 日志工具
├── web/
│   ├── templates/
│   │   └── index.html            # Web UI模板
│   └── static/
│       ├── css/
│       │   └── style.css         # 样式文件
│       └── js/
│           └── app.js            # 前端逻辑
├── scripts/
│   ├── init.sh                   # Linux/Mac初始化脚本
│   └── init.ps1                  # Windows初始化脚本
├── docker-compose.yml            # Docker Compose配置
├── Dockerfile                    # Docker镜像构建文件
├── Makefile                      # Make命令
├── go.mod                        # Go模块定义
├── go.sum                        # Go依赖校验
├── .gitignore                    # Git忽略文件
├── README.md                     # 项目说明
└── PROJECT.md                    # 项目结构说明（本文件）
```

## 🔧 核心功能实现

### 1. 短链接生成
- **位置**: `services/link_service.go`
- **功能**: 
  - 生成随机短代码（支持自定义）
  - 动态长度扩展（6位→7位→...）
  - 内容哈希一致性检查

### 2. 数据库层
- **位置**: `database/`
- **功能**:
  - PostgreSQL连接管理
  - 自动数据库迁移
  - CRUD操作封装

### 3. 搜索功能
- **位置**: `services/search_service.go`
- **功能**:
  - Meilisearch集成
  - 全文搜索支持
  - 自动索引更新

### 4. API接口
- **位置**: `handlers/`
- **功能**:
  - RESTful API设计
  - Token认证
  - 错误处理

### 5. Web UI
- **位置**: `web/`
- **功能**:
  - 响应式设计
  - 链接管理
  - 统计展示
  - 搜索功能

## 🚀 部署说明

### 开发环境
1. 安装Go 1.21+
2. 安装PostgreSQL和Meilisearch
3. 配置环境变量
4. 运行 `go run cmd/server/main.go`

### 生产环境
1. 使用Docker Compose一键部署
2. 配置环境变量
3. 运行 `docker-compose up -d`

## 📝 开发注意事项

1. **日志系统**: 所有日志通过 `utils/logger.go` 统一管理
2. **配置管理**: 配置通过环境变量和数据库双重管理
3. **错误处理**: 所有错误都有适当的日志记录
4. **代码注释**: 所有文件和函数都有中文注释

## 🔐 安全特性

1. API Token认证
2. SQL注入防护（使用参数化查询）
3. 输入验证
4. 错误信息不泄露敏感信息

## 📊 性能优化

1. 数据库索引优化
2. Meilisearch全文搜索
3. 连接池管理
4. 缓存策略（可扩展）

## 🔄 扩展建议

1. **Redis缓存**: 可以添加Redis缓存热门链接
2. **限流**: 可以添加API限流功能
3. **多租户**: 可以扩展支持多用户/多租户
4. **批量操作**: 可以添加批量创建/删除功能
5. **导出功能**: 可以添加数据导出功能
6. **API文档**: 可以集成Swagger/OpenAPI文档

