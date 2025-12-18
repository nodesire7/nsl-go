/**
 * v2 路由注册（重写版）
 * 将新架构的 handler/middleware 挂载到 gin.Router
 */
package httpv2

import (
	"context"
	"fmt"
	"time"

	"short-link/internal/config"
	"short-link/internal/db"
	"short-link/internal/httpv2/handlers"
	v2mw "short-link/internal/httpv2/middleware"
	"short-link/internal/repo"
	"short-link/internal/service"
	"short-link/middleware"
	"short-link/utils"

	"github.com/gin-gonic/gin"
)

// Module v2 模块（重写版）
type Module struct {
	Cfg         *config.Config
	Pool        *db.Pool
	UserRepo    *repo.UserRepo
	DomainRepo  *repo.DomainRepo
	SettingsRepo *repo.SettingsRepo
	LinkRepo    *repo.LinkRepo
	AccessLogRepo *repo.AccessLogRepo
	UserService *service.UserService
	LinkService *service.LinkService
	SearchService *service.SearchService
	AuthHandler *handlers.AuthHandler
	LinkHandler *handlers.LinkHandler
	RedirectHandler *handlers.RedirectHandler
	StatsHandler *handlers.StatsHandler
}

// New 创建 v2 模块
func New() (*Module, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := db.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化pgxpool失败: %w", err)
	}

	// 执行版本化迁移（internal/db/migrations/*.sql）
	if err := db.Migrate(ctx, pool); err != nil {
		return nil, fmt.Errorf("执行数据库迁移失败: %w", err)
	}

	userRepo := repo.NewUserRepo(pool)
	domainRepo := repo.NewDomainRepo(pool)
	settingsRepo := repo.NewSettingsRepo(pool)
	linkRepo := repo.NewLinkRepo(pool)
	accessLogRepo := repo.NewAccessLogRepo(pool)

	userService := service.NewUserService(userRepo)
	linkService := service.NewLinkService(cfg.BaseURL, cfg.MinCodeLength, cfg.MaxCodeLength, linkRepo, domainRepo, settingsRepo, userRepo, accessLogRepo)
	searchService, err := service.NewSearchService(cfg)
	if err != nil {
		utils.LogWarn("Meilisearch(v2) 初始化失败，搜索功能将不可用: %v", err)
		searchService = nil
	}

	authHandler := handlers.NewAuthHandler(cfg, userService)
	linkHandler := handlers.NewLinkHandler(cfg, linkService, linkRepo, domainRepo, searchService)
	redirectHandler := handlers.NewRedirectHandler(linkService)
	statsHandler := handlers.NewStatsHandler(linkService)

	return &Module{
		Cfg:         cfg,
		Pool:        pool,
		UserRepo:    userRepo,
		DomainRepo:  domainRepo,
		SettingsRepo: settingsRepo,
		LinkRepo:    linkRepo,
		AccessLogRepo: accessLogRepo,
		UserService: userService,
		LinkService: linkService,
		SearchService: searchService,
		AuthHandler: authHandler,
		LinkHandler: linkHandler,
		RedirectHandler: redirectHandler,
		StatsHandler: statsHandler,
	}, nil
}

// Close 关闭资源
func (m *Module) Close() {
	if m != nil && m.Pool != nil {
		m.Pool.Close()
	}
}

// RegisterRoutes 注册 v2 路由
func RegisterRoutes(router *gin.Engine, m *Module) {
	utils.LogInfo("挂载重写版路由：/api/v2")

	// 重写版 redirect（替换 legacy 的任意域名查询，修复多域名 code 冲突风险）
	router.GET("/:code", m.RedirectHandler.Redirect)

	api := router.Group("/api/v2")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", m.AuthHandler.Register)
			authGroup.POST("/login", m.AuthHandler.Login)
			authGroup.POST("/logout", m.AuthHandler.Logout)
		}

		protected := api.Group("")
		protected.Use(v2mw.AuthMiddleware(m.Cfg.JWTSecret, m.UserRepo))
		protected.Use(middleware.CSRFMiddleware())
		{
			protected.GET("/profile", m.AuthHandler.GetProfile)
			protected.POST("/profile/token", m.AuthHandler.UpdateToken)

			// 链接管理（v2 优先迁移核心能力：创建/列表）
			protected.POST("/links", m.LinkHandler.CreateLink)
			protected.GET("/links", m.LinkHandler.GetLinks)
			protected.GET("/links/search", m.LinkHandler.SearchLinks)
			protected.DELETE("/links/:code", m.LinkHandler.DeleteLink)

			// 统计
			protected.GET("/stats", m.StatsHandler.GetStats)
		}
	}
}


