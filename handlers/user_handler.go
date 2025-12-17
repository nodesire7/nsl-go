/**
 * 用户处理器
 * 处理用户注册、登录等HTTP请求
 */
package handlers

import (
	"net/http"
	"short-link/models"
	"short-link/services"
	"short-link/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler 创建用户处理器实例
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}
	
	user, err := h.userService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成token失败",
		})
		return
	}

	// 下发 Cookie（HttpOnly）+ CSRF Cookie（双提交）
	csrfToken, _ := utils.GenerateCSRFToken()
	c.SetSameSite(http.SameSiteLaxMode)
	secure := c.Request.TLS != nil
	// access_token：HttpOnly
	c.SetCookie("access_token", token, int((24*time.Hour).Seconds()), "/", "", secure, true)
	// csrf_token：前端可读
	c.SetCookie("csrf_token", csrfToken, int((24*time.Hour).Seconds()), "/", "", secure, false)
	
	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token, // 兼容：仍返回 token（API 客户端可用）
		User: models.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			APIToken:  user.APIToken, // 重写版建议仅在注册/重置时展示；现阶段保持兼容
			Role:      user.Role,
			MaxLinks:  user.MaxLinks,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05"),
		},
	})
}

// UpdateToken 更新用户Token
func (h *UserHandler) UpdateToken(c *gin.Context) {
	userID := c.GetInt64("user_id")
	
	newToken, err := h.userService.UpdateUserToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新Token失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"api_token": newToken,
		"message": "Token已更新，旧Token已失效",
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}
	
	user, err := h.userService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成token失败",
		})
		return
	}

	// 下发 Cookie（HttpOnly）+ CSRF Cookie（双提交）
	csrfToken, _ := utils.GenerateCSRFToken()
	c.SetSameSite(http.SameSiteLaxMode)
	secure := c.Request.TLS != nil
	c.SetCookie("access_token", token, int((24*time.Hour).Seconds()), "/", "", secure, true)
	c.SetCookie("csrf_token", csrfToken, int((24*time.Hour).Seconds()), "/", "", secure, false)
	
	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token, // 兼容：仍返回 token（API 客户端可用）
		User: models.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			APIToken:  "", // 安全：不在登录接口回传长期API Token（如需请调用 /api/v1/profile/token 生成新token）
			Role:      user.Role,
			MaxLinks:  user.MaxLinks,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05"),
		},
	})
}

// Logout 退出登录（清理Cookie）
func (h *UserHandler) Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	secure := c.Request.TLS != nil
	// MaxAge=-1 删除Cookie
	c.SetCookie("access_token", "", -1, "/", "", secure, true)
	c.SetCookie("csrf_token", "", -1, "/", "", secure, false)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetProfile 获取用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64("user_id")
	
	user, err := h.userService.GetUserInfo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取用户信息失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.UserInfo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		APIToken:  "", // 安全：profile 不返回长期API Token
		Role:      user.Role,
		MaxLinks:  user.MaxLinks,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05"),
	})
}

