/**
 * 主程序入口
 * 启动HTTP服务器和初始化所有服务
 */
package main

import (
	"fmt"
	"log"
	"short-link/cache"
	"short-link/config"
	"short-link/database"
	"short-link/handlers"
	"short-link/middleware"
	"short-link/services"
	"short-link/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	log.Printf("配置加载完成，服务端口: %d", cfg.ServerPort)
	
	// 初始化日志
	utils.InitLogger()
	utils.LogInfo("日志系统初始化完成")
	
	// 初始化JWT
	utils.InitJWT()
	utils.LogInfo("JWT初始化完成")
	
	// 初始化数据库
	if err := database.InitDB(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.CloseDB()
	utils.LogInfo("数据库初始化完成")
	
	// 初始化admin用户
	if err := database.InitAdminUser(); err != nil {
		utils.LogWarn("Admin用户初始化失败: %v", err)
	}
	
	// 初始化Redis
	if err := cache.InitRedis(); err != nil {
		utils.LogWarn("Redis初始化失败，缓存功能将不可用: %v", err)
	}
	defer cache.CloseRedis()
	
	// 初始化搜索服务
	var searchService *services.SearchService
	var err error
	searchService, err = services.NewSearchService()
	if err != nil {
		utils.LogWarn("Meilisearch初始化失败，搜索功能将不可用: %v", err)
		searchService = nil
	} else {
		utils.LogInfo("Meilisearch初始化完成")
	}
	
	// 初始化服务
	linkService := services.NewLinkService()
	userService := services.NewUserService()
	domainService := services.NewDomainService()
	utils.LogInfo("服务初始化完成")
	
	// 初始化处理器
	linkHandler := handlers.NewLinkHandler(linkService, searchService, userService, domainService)
	statsHandler := handlers.NewStatsHandler(linkService)
	settingsHandler := handlers.NewSettingsHandler()
	userHandler := handlers.NewUserHandler(userService)
	domainHandler := handlers.NewDomainHandler(domainService)
	
	// 设置Gin模式
	if cfg.ServerMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	// 创建Gin路由
	router := gin.Default()
	
	// 中间件
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RateLimitMiddleware())
	
	// 静态文件服务（用于Web UI）
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")
	
	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"service": "short-link",
		})
	})
	
	// 重定向路由（不需要认证）
	router.GET("/:code", linkHandler.RedirectLink)
	
	// Web UI路由（页面本身不强制要求Authorization头；由前端携带JWT访问API）
	router.GET("/login", func(c *gin.Context) {
		c.HTML(200, "login.html", gin.H{
			"title": "登录 - 短链接管理系统",
		})
	})
	router.GET("/register", func(c *gin.Context) {
		c.HTML(200, "register.html", gin.H{
			"title": "注册 - 短链接管理系统",
		})
	})
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", gin.H{
			"title": "短链接管理系统",
		})
	})
	
	// API路由
	api := router.Group("/api/v1")
	{
		// 认证（不需要JWT认证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}
		
		// 需要认证的API
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// 用户相关
			protected.GET("/profile", userHandler.GetProfile)
			protected.POST("/profile/token", userHandler.UpdateToken) // 更新用户Token
			
			// 域名管理
			protected.POST("/domains", domainHandler.CreateDomain)
			protected.GET("/domains", domainHandler.GetDomains)
			protected.DELETE("/domains/:id", domainHandler.DeleteDomain)
			protected.PUT("/domains/:id/default", domainHandler.SetDefaultDomain)
			
			// 链接管理
			protected.POST("/links", linkHandler.CreateLink)
			protected.GET("/links", linkHandler.GetLinks)
			protected.GET("/links/search", linkHandler.SearchLinks)
			protected.GET("/links/:code", linkHandler.GetLinkInfo)
			protected.DELETE("/links/:code", linkHandler.DeleteLink)
			
			// 统计
			protected.GET("/stats", statsHandler.GetStats)
			
			// 配置管理
			protected.GET("/settings", settingsHandler.GetSettings)
			protected.PUT("/settings", settingsHandler.UpdateSettings)
		}
	}
	
	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	utils.LogInfo("服务器启动在 %s", addr)
	
	if err := router.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

