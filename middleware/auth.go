/**
 * 认证中间件
 * 提供JWT Token和API Token认证功能
 */
package middleware

import (
	"net/http"
	"short-link/database"
	"short-link/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT或API Token认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过健康检查、注册、登录端点（仅用于API/页面可访问性）
		if c.Request.URL.Path == "/health" || 
		   c.Request.URL.Path == "/login" ||
		   c.Request.URL.Path == "/register" ||
		   c.Request.URL.Path == "/api/v1/auth/register" ||
		   c.Request.URL.Path == "/api/v1/auth/login" {
			c.Next()
			return
		}
		
		// 跳过静态文件和重定向
		if strings.HasPrefix(c.Request.URL.Path, "/static/") ||
		   strings.HasPrefix(c.Request.URL.Path, "/api/v1/redirect/") {
			c.Next()
			return
		}
		
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "缺少认证令牌",
			})
			c.Abort()
			return
		}
		
		// 支持 Bearer token 格式
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		
		// 先尝试JWT认证
		claims, err := utils.ParseToken(token)
		if err == nil && claims != nil {
			// JWT认证成功
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Next()
			return
		}
		
		// 尝试用户API Token认证
		user, err := database.GetUserByToken(token)
		if err == nil && user != nil {
			// 用户Token认证成功
			c.Set("user_id", user.ID)
			c.Set("username", user.Username)
			c.Set("role", user.Role)
			c.Next()
			return
		}
		
		// 注意：不再允许 API_TOKEN 作为“超级管理员通行证”绕过所有权限。
		// 如果确实需要系统级token，应在重写版中限定到特定管理端点，并进行审计与RBAC控制。
		
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "无效的认证令牌",
		})
		c.Abort()
	}
}

