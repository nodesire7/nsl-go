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
	
	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User: models.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			APIToken:  user.APIToken,
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
	
	c.JSON(http.StatusOK, models.LoginResponse{
		Token: token,
		User: models.UserInfo{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			APIToken:  user.APIToken,
			Role:      user.Role,
			MaxLinks:  user.MaxLinks,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05"),
		},
	})
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
		APIToken:  user.APIToken,
		Role:      user.Role,
		MaxLinks:  user.MaxLinks,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05"),
	})
}

