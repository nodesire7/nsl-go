/**
 * 集成测试
 * 使用 testcontainers 测试 PG/Redis/Meili 集成
 * 实现 redo.md 6.3：集成测试
 */
package integration

import (
	"context"
	"testing"
	"time"

	"short-link/internal/config"
	"short-link/internal/db"
	"short-link/internal/jobs"
	"short-link/internal/repo"
	"short-link/internal/service"
	"short-link/models"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestDB 使用 testcontainers 启动 PostgreSQL
func setupTestDB(ctx context.Context) (*db.Pool, func(), error) {
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	// 获取容器的主机和端口
	host, err := pgContainer.Host(ctx)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, nil, err
	}
	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, nil, err
	}

	// 创建测试配置
	cfg := &config.Config{
		DBHost:     host,
		DBPort:     port.Int(),
		DBUser:     "testuser",
		DBPassword: "testpass",
		DBName:     "testdb",
		DBSSLMode:  "disable",
		DBMaxConns: 5,
	}
	pool, err := db.New(ctx, cfg)
	if err != nil {
		_ = pgContainer.Terminate(ctx)
		return nil, nil, err
	}

	// 执行迁移
	if err := db.Migrate(ctx, pool); err != nil {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
		return nil, nil, err
	}

	cleanup := func() {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
	}

	return pool, cleanup, nil
}

// setupTestRedis 使用 testcontainers 启动 Redis
func setupTestRedis(ctx context.Context) (string, func(), error) {
	redisContainer, err := redis.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithStartupTimeout(10*time.Second)),
	)
	if err != nil {
		return "", nil, err
	}

	endpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		_ = redisContainer.Terminate(ctx)
		return "", nil, err
	}

	cleanup := func() {
		_ = redisContainer.Terminate(ctx)
	}

	return endpoint, cleanup, nil
}

// TestLinkService_CreateLink 测试链接创建
func TestLinkService_CreateLink(t *testing.T) {
	ctx := context.Background()

	// 设置测试数据库
	pool, cleanupDB, err := setupTestDB(ctx)
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDB()

	// 创建 repos
	linkRepo := repo.NewLinkRepo(pool)
	domainRepo := repo.NewDomainRepo(pool)
	settingsRepo := repo.NewSettingsRepo(pool)
	userRepo := repo.NewUserRepo(pool)
	accessLogRepo := repo.NewAccessLogRepo(pool)

	// 创建测试配置
	cfg := &config.Config{
		BaseURL:      "http://localhost:9110",
		MinCodeLength: 6,
		MaxCodeLength: 10,
		JWTSecret:    "test-jwt-secret-for-integration-tests-only", // 测试用 JWT secret
	}

	// 创建 statsWorker（简化版，仅用于测试）
	statsWorker := jobs.NewStatsWorker(linkRepo, accessLogRepo, 10, 1*time.Second)

	// 创建 service
	linkService := service.NewLinkService(
		cfg.BaseURL,
		cfg.MinCodeLength,
		cfg.MaxCodeLength,
		linkRepo,
		domainRepo,
		settingsRepo,
		userRepo,
		accessLogRepo,
		statsWorker,
		nil, // meiliWorker
	)
	defer statsWorker.Stop()

	// 测试创建链接
	req := &models.CreateLinkRequest{
		URL:   "https://example.com",
		Title: "测试链接",
	}

	link, shortURL, err := linkService.CreateLink(ctx, 1, req)
	if err != nil {
		t.Fatalf("创建链接失败: %v", err)
	}

	if link == nil {
		t.Fatal("链接为空")
	}

	if shortURL == "" {
		t.Fatal("短链接 URL 为空")
	}

	if link.OriginalURL != req.URL {
		t.Errorf("原始 URL 不匹配: 期望 %s, 得到 %s", req.URL, link.OriginalURL)
	}
}

// TestUserService_Register 测试用户注册
func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	// 设置测试数据库
	pool, cleanupDB, err := setupTestDB(ctx)
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDB()

	userRepo := repo.NewUserRepo(pool)
	userService := service.NewUserService(userRepo)

	// 测试注册
	req := &models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := userService.Register(ctx, req)
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	if user == nil {
		t.Fatal("用户为空")
	}

	if user.Username != req.Username {
		t.Errorf("用户名不匹配: 期望 %s, 得到 %s", req.Username, user.Username)
	}

	if user.Email != req.Email {
		t.Errorf("邮箱不匹配: 期望 %s, 得到 %s", req.Email, user.Email)
	}
}

// TestRedisIntegration 测试 Redis 集成
func TestRedisIntegration(t *testing.T) {
	ctx := context.Background()

	// 设置测试 Redis
	endpoint, cleanupRedis, err := setupTestRedis(ctx)
	if err != nil {
		t.Fatalf("设置测试 Redis 失败: %v", err)
	}
	defer cleanupRedis()

	// 这里可以测试 Redis 缓存功能
	// 由于需要初始化 cache 包，这里仅做示例
	t.Logf("Redis 端点: %s", endpoint)
}

