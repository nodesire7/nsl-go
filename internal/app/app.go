/**
 * 应用启动封装
 * 应用启动封装（重写版入口）
 * - 仅使用 internal/* + v2 路由，不再挂载 /api/v1
 */
package app

import (
	"fmt"
	"short-link/cache"
	icfg "short-link/internal/config"
	"short-link/internal/httpv2"
	"short-link/internal/tracing"
	"short-link/middleware"
	"short-link/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Run 启动 HTTP 服务
func Run() error {
	// 初始化日志
	utils.InitLogger()
	utils.LogInfo("日志系统初始化完成")

	// 加载重写版配置（internal/config）
	cfg, err := icfg.Load()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}
	utils.LogInfo("配置加载完成，服务端口: %d", cfg.ServerPort)

	// 初始化Redis
	if err := cache.InitRedis(); err != nil {
		utils.LogWarn("Redis初始化失败，缓存功能将不可用: %v", err)
	}
	defer func() {
		if err := cache.CloseRedis(); err != nil {
			utils.LogWarn("关闭Redis连接失败: %v", err)
		}
	}()
	
	// 初始化限流器（滑动窗口 + 令牌桶）
	middleware.InitRateLimiters()

	// 初始化 Tracing（可选）
	cleanupTracing, err := tracing.InitTracing(cfg)
	if err != nil {
		utils.LogWarn("Tracing 初始化失败: %v", err)
	} else {
		defer cleanupTracing()
		if cfg.JaegerEndpoint != "" {
			utils.LogInfo("Tracing 已启用（OTLP: %s）", cfg.JaegerEndpoint)
		}
	}

	// 设置Gin模式
	if cfg.ServerMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由
	router := gin.Default()

	// 中间件（全局）
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.MetricsMiddleware()) // Prometheus 指标收集
	
	// Tracing 中间件（如果启用）
	if cfg.JaegerEndpoint != "" {
		router.Use(otelgin.Middleware("nsl-go"))
	}
	
	router.Use(middleware.RateLimitMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.RequestIDMiddleware())

	// 静态文件服务（用于Web UI）
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "short-link",
		})
	})

	// Prometheus metrics 端点
	router.GET("/metrics", func(c *gin.Context) {
		// 使用 Prometheus 的 HTTP handler
		handler := promhttp.Handler()
		handler.ServeHTTP(c.Writer, c.Request)
	})

	// Web UI 路由
	router.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", gin.H{"title": "登录 - 短链接管理系统"})
	})
	router.GET("/register", func(c *gin.Context) {
		c.HTML(200, "register.html", gin.H{"title": "注册 - 短链接管理系统"})
	})
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{"title": "短链接管理系统"})
	})

	// 挂载重写版 v2 路由（现在作为唯一 API 版本）
	if v2, err := httpv2.New(); err != nil {
		utils.LogError("v2模块初始化失败: %v", err)
		return err
	} else {
		defer v2.Close()
		// 启动异步统计 Worker
		if v2.StatsWorker != nil {
			v2.StatsWorker.Start()
		}
		// 启动 Meilisearch Worker
		if v2.LinkService != nil && v2.LinkService.GetMeiliWorker() != nil {
			v2.LinkService.GetMeiliWorker().Start()
		}
		httpv2.RegisterRoutes(router, v2)
	}

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	utils.LogInfo("服务器启动在 %s", addr)
	return router.Run(addr)
}


