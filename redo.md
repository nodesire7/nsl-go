# 🔁 重写计划（Redo）——New short link (NSL GO)

> 本文用于：总结现状问题、明确安全基线与新架构，并作为“重写版本”的实施蓝图与验收标准。

---

## 1. 背景与目标

### 1.1 背景
当前实现已具备“短链生成/跳转、PG、Meilisearch、用户系统、多域名、二维码、限流、CI/CD”等功能，但在**安全性、可维护性、并发正确性、可观测性、可测试性**上存在明显不足，需要进行系统性重写。

### 1.2 重写目标（必须达成）
- **安全**：默认安全（secure by default），杜绝“单个配置 token 绕过全部权限”的设计。
- **正确性**：并发下短码生成无竞态、无重复、可回退重试；统计写入稳定。
- **性能**：热点跳转读路径可缓存；高并发下 DB/Redis/Meili 均有超时、限流与降级策略。
- **可维护**：分层清晰（handler/service/repo），依赖注入、接口隔离，可做单测与集成测。
- **可观测**：结构化日志、指标、追踪、审计日志（尤其是管理员操作）。
- **交付**：保留现有 API 大体结构，同时提供明确的迁移说明。

---

## 2. 现状问题清单（含你朋友反馈 + 进一步补充）

### 2.1 并发与资源管理
- **全局 DB 句柄的误解**：Go 的 `*sql.DB` 本身线程安全，内部自带连接池；但当前未配置池参数（`MaxOpen/MaxIdle/ConnMaxLifetime`），在高并发可能导致连接抖动或资源占用失控。
- **缺少 context 超时**：DB/Redis/Meili 请求缺少 `context.WithTimeout`，慢查询或网络抖动会造成 goroutine 堆积。
- **全局 context**：`context.Background()` 全局复用不便于取消/超时控制；不算“泄漏”，但会导致**无法收敛资源**。

### 2.2 随机数与短码生成
- **使用 `math/rand.Seed(time.Now().UnixNano())`**：可预测/可重复，且每次调用都 Seed 会降低随机质量与产生碰撞风险（并发时更明显）。
- **短码生成竞态**：即使使用更好的随机数，仍需要在 DB 层用唯一约束 + 冲突重试，避免并发插入重复 code。
- **“长度耗尽”计算成本**：按长度统计、计算 90% 阈值可能导致高频 COUNT 查询；需要更高效策略（例如：只在冲突频率升高时扩容）。

### 2.3 鉴权/授权（严重）
- **JWT 密钥硬编码/与 API_TOKEN 耦合**：存在默认密钥、或直接复用 API_TOKEN 的风险，导致 token 可伪造。
- **API_TOKEN “超级通行证”**：若允许某个系统 token 直接获得 admin 权限，会导致：
  - 任何泄漏都等价于“全站失守”
  - 无审计、无最小权限
  - 与用户 token/JWT 体系冲突
- **Token 存储明文**：用户 API token 直接存数据库明文字段；一旦数据库泄漏，所有 token 可被直接利用。应改为存储 hash（如 SHA-256/argon2）。
- **Web 端 token 存储策略不安全**：localStorage 容易被 XSS 窃取。更推荐 HttpOnly Cookie + CSRF 防护，或 BFF 模式。

### 2.4 Web 安全
- **缺少 CSRF 防护**：若采用 Cookie 鉴权（未来大概率），必须加 CSRF token 或 SameSite 策略。
- **缺少基础安全头**：CSP、HSTS、X-Frame-Options、X-Content-Type-Options 等。
- **URL 安全校验不足**：可能被用于 SSRF、内网探测、跳转到恶意协议（`file://`、`javascript:` 等）。

### 2.5 Redis 的必要性与使用方式
- 当前 Redis 主要用于**限流计数**，并未用于热点跳转缓存；对“高并发短链跳转”场景帮助有限。
- 限流实现是简单计数器：缺少更精确的滑动窗口/令牌桶策略、缺少代理链路真实 IP 处理策略。

### 2.6 Meilisearch 同步与一致性
- 创建/删除/更新短链时对 Meili 的写入失败如何处理（重试、补偿、死信、后台任务）不明确。

### 2.7 日志与审计
- 目前日志混用 `log.Printf` 与自定义 logger，缺少 request_id、用户信息、耗时、错误堆栈。
- **敏感信息输出**：首次创建 admin 时输出明文密码/Token 存在泄漏风险（日志系统、容器日志平台、CI 采集都会暴露）。

### 2.8 测试与质量
- 缺少单元测试、集成测试（PG/Redis/Meili）、安全测试用例（SSRF/XSS/CSRF/权限绕过）。
- 缺少 lint（`golangci-lint`）与安全扫描（`gosec`）规范化配置。

---

## 3. 新架构总览（重写版）

### 3.1 目录结构（建议）
```
nsl-go/
  cmd/
    api/            # HTTP 服务入口
    nsl/            # CLI 管理工具（admin/迁移/修复/导出）
  internal/
    config/         # 配置加载（env + file），强校验
    http/           # gin/chi 路由、middleware、handler
    service/        # 业务逻辑（短链/用户/域名/统计/搜索）
    repo/           # 数据访问层（pgxpool + sqlc / bun）
    auth/           # JWT、API key、RBAC、session
    cache/          # Redis 缓存（热点跳转、限流、会话）
    search/         # Meilisearch 适配层
    jobs/           # 后台任务（索引重建、统计聚合、补偿重试）
    observability/  # logger/metrics/tracing
  web/              # Web UI 静态资源（可改为独立前端项目）
  migrations/       # 版本化迁移（golang-migrate）
  docs/
    security.md
    api.md
    deployment.md
```

### 3.2 分层原则
- **Handler**：只做参数校验、鉴权上下文读取、调用 service、返回响应。
- **Service**：业务逻辑（幂等、冲突重试、权限校验、缓存策略、任务派发）。
- **Repo**：纯数据访问（必须全程使用 context + timeout）。
- **Middleware**：鉴权、限流、日志、CORS、安全头。

---

## 4. 安全基线（必须）

### 4.1 密钥与配置
- `JWT_SECRET` **必须显式配置**（生产环境），不允许硬编码默认值。
- API Token **不再作为超级通行证**：若保留系统级 token，只能用于少量管理端点，并且必须审计 + RBAC。
- 用户 API token 只存 **hash**（例如 `sha256(token)`），只在创建时展示一次。

### 4.2 鉴权模型（建议方案）
- **Web UI**：HttpOnly Cookie + SameSite + CSRF token（双提交或服务端存储）。
- **API 调用**：
  - Header Bearer API Token（用户级、永久），或
  - JWT（短期 access + 可选 refresh）。
- **RBAC**：
  - 基础角色：`admin` / `user`
  - 预留权限点：`link:create` `link:delete` `domain:manage` `settings:update` 等

### 4.3 输入与URL安全
- 仅允许 `http/https` scheme
- SSRF 防护：禁止内网 IP / link-local / loopback / metadata endpoint（可配置开关）
- URL 规范化后再做 hash（避免“同一URL不同写法”重复入库）

---

## 5. 核心业务重写要点

### 5.1 短码生成（并发安全）
- 生成：`crypto/rand` + base62（拒绝采样）
- 插入：依赖数据库唯一约束（`(domain_id, code)`）冲突后重试
- 动态长度：
  - 默认最小 6 位
  - 当冲突率/容量阈值到达时自动扩到 7、8…
  - 管理员可设 `min_len`/`max_len`

### 5.2 幂等：URL 内容 hash 一致性
- `hash = sha256(normalized_url)`
- 唯一约束建议：`(user_id, hash)` 或 `(user_id, domain_id, hash)`（按产品选择）
- 再提交相同 URL：返回已有短链（包括二维码）

### 5.3 统计与访问日志
- 跳转路径必须极快：
  - **读：Redis 缓存 code -> url + link_id**
  - **写：异步写入访问日志/点击数**（队列/worker/batch）
- 提供聚合统计：日/周/月、TopN、来源、UA 等

### 5.4 Redis 的定位
- 必须用途：热点跳转缓存、限流、会话/CSRF token（若使用 Cookie）
- 可选用途：短链详情缓存、用户配额计数缓存、Meili 写入失败队列

---

## 6. 交付与验收标准

### 6.1 功能验收
- 短链生成/跳转正常
- 相同 URL 幂等返回同一短链
- 动态长度扩容可用（可配置 min/max）
- 多域名增删改查可用
- 二维码生成可用
- 统计与访问记录可用
- Web UI：登录页、管理页可用
- CLI：初始化 admin、重置密码、重置 token、导出数据等

### 6.2 安全验收（必测）
- 无硬编码密钥
- 无“系统 token 全权限绕过”
- token 不明文存储
- CSRF（若 cookie）通过
- SSRF 拦截通过
- RBAC 权限点可扩展

### 6.3 质量验收
- 单元测试覆盖关键 service（短码生成、幂等、权限校验、配额）
- 集成测试：PG/Redis/Meili（建议 testcontainers）
- CI：`go test ./...` + `golangci-lint` + `gosec`
- 结构化日志 + request_id + metrics（Prometheus）

---

## 7. 重写实施路线（建议）

1) **先定安全与配置**：JWT_SECRET、token hash、移除超级通行证、审计日志
   - ✅ JWT_SECRET、token hash、移除超级通行证
   - ❌ 审计日志（管理员操作、敏感操作记录）

2) **重写数据层**：pgxpool + 迁移版本化（migrate）
   - ✅ 已完成

3) **重写鉴权与RBAC**：web cookie + csrf / api token hash
   - ✅ web cookie + csrf / api token hash
   - ❌ RBAC 权限点（目前仅角色字段，缺少细粒度权限点）

4) **重写短码生成与幂等**：并发安全 + DB 冲突重试
   - ✅ 已完成

5) **引入缓存与异步统计**：热点缓存 + worker
   - ✅ 已完成

6) **补齐测试与CI**：单测/集成测/安全扫描
   - ❌ 集成测试（PG/Redis/Meili）
   - ❌ `golangci-lint` / `gosec` 在 CI 中落地
   - ❌ Metrics（Prometheus）
   - ❌ Tracing

---

## 8. 需要你确认的产品决策（重写前必须定）
- **幂等的粒度**：同一用户同一 URL 是否允许多个 domain 下生成不同短链？ 这里允许
- **Web UI 鉴权方式**：Cookie（更安全）还是 localStorage（更简单但有XSS风险）？ 这里使用Cookie（更安全） 
- **系统级 API_TOKEN 是否保留**：如果保留，允许哪些端点？是否强制IP白名单/二次验证？ 不保留


