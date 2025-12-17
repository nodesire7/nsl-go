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
	UserService *service.UserService
	AuthHandler *handlers.AuthHandler
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

	userRepo := repo.NewUserRepo(pool)
	userService := service.NewUserService(userRepo)
	authHandler := handlers.NewAuthHandler(cfg, userService)

	return &Module{
		Cfg:         cfg,
		Pool:        pool,
		UserRepo:    userRepo,
		UserService: userService,
		AuthHandler: authHandler,
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
		}
	}
}


