/**
 * v2 用户认证 Handler（重写版）
 * /api/v2/auth/register
 * /api/v2/auth/login
 * /api/v2/auth/logout
 * /api/v2/profile
 * /api/v2/profile/token
 */
package handlers

import (
	"context"
	"net/http"
	"time"

	"short-link/internal/auth"
	appcfg "short-link/internal/config"
	"short-link/internal/repo"
	"short-link/internal/service"
	"short-link/models"
	"short-link/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler v2 认证处理器
type AuthHandler struct {
	cfg         *appcfg.Config
	userService *service.UserService
	auditLogRepo *repo.AuditLogRepo
}

// NewAuthHandler 创建 AuthHandler
func NewAuthHandler(cfg *appcfg.Config, userService *service.UserService, auditLogRepo *repo.AuditLogRepo) *AuthHandler {
	return &AuthHandler{cfg: cfg, userService: userService, auditLogRepo: auditLogRepo}
}

// Register 注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	u, err := h.userService.Register(ctx, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtToken, err := auth.GenerateJWT(h.cfg.JWTSecret, u.ID, u.Username, u.Role, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	// Cookie（HttpOnly）+ CSRF Cookie（双提交）
	csrfToken, _ := utils.GenerateCSRFToken()
	c.SetSameSite(http.SameSiteLaxMode)
	secure := c.Request.TLS != nil
	c.SetCookie("access_token", jwtToken, int((24*time.Hour).Seconds()), "/", "", secure, true)
	c.SetCookie("csrf_token", csrfToken, int((24*time.Hour).Seconds()), "/", "", secure, false)

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: jwtToken,
		User: models.UserInfo{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			APIToken:  u.APIToken, // 仅注册返回一次，便于 API 客户端保存
			Role:      u.Role,
			MaxLinks:  u.MaxLinks,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05"),
		},
	})
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	u, err := h.userService.Login(ctx, &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	jwtToken, err := auth.GenerateJWT(h.cfg.JWTSecret, u.ID, u.Username, u.Role, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成token失败"})
		return
	}

	csrfToken, _ := utils.GenerateCSRFToken()
	c.SetSameSite(http.SameSiteLaxMode)
	secure := c.Request.TLS != nil
	c.SetCookie("access_token", jwtToken, int((24*time.Hour).Seconds()), "/", "", secure, true)
	c.SetCookie("csrf_token", csrfToken, int((24*time.Hour).Seconds()), "/", "", secure, false)

	c.JSON(http.StatusOK, models.LoginResponse{
		Token: jwtToken,
		User: models.UserInfo{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			APIToken:  "", // 安全：不在登录回传长期 API Token
			Role:      u.Role,
			MaxLinks:  u.MaxLinks,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05"),
		},
	})
}

// Logout 退出登录
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	secure := c.Request.TLS != nil
	c.SetCookie("access_token", "", -1, "/", "", secure, true)
	c.SetCookie("csrf_token", "", -1, "/", "", secure, false)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetProfile 获取当前用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	u, err := h.userService.GetUserInfo(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.UserInfo{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		APIToken:  "", // 安全：profile 不返回长期 API Token
		Role:      u.Role,
		MaxLinks:  u.MaxLinks,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05"),
	})
}

// UpdateToken 轮换 API Token
func (h *AuthHandler) UpdateToken(c *gin.Context) {
	userID := c.GetInt64("user_id")
	username := c.GetString("username")
	role := c.GetString("role")
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	newToken, err := h.userService.RotateAPIToken(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新Token失败: " + err.Error()})
		return
	}

	// 记录审计日志（敏感操作）
	if h.auditLogRepo != nil {
		auditLog := &models.AuditLog{
			UserID:       &userID,
			Username:     username,
			Action:       "token.rotate",
			ResourceType: "user",
			ResourceID:   &userID,
			IP:           c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
			Details: map[string]interface{}{
				"role": role,
			},
			CreatedAt: time.Now(),
		}
		_ = h.auditLogRepo.CreateAuditLog(ctx, auditLog) // best-effort
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"api_token": newToken,
		"message":   "Token已更新，旧Token已失效",
	})
}


