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
		// 跳过健康检查、注册、登录端点（页面和公开API）
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
		
		// 认证来源优先级：
		// 1) Authorization: Bearer <token>（API Token 或 JWT）
		// 2) Cookie: access_token（Web UI 登录）
		authHeader := c.GetHeader("Authorization")
		token := ""
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Cookie JWT（HttpOnly）
			if cookie, err := c.Cookie("access_token"); err == nil {
				token = cookie
			}
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
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
		
		// 尝试用户API Token认证（仅当显式通过 Authorization 传入时生效）
		// 避免把 API token 放进 Cookie 触发 CSRF 风险。
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			user, err := database.GetUserByToken(token)
			if err == nil && user != nil {
				c.Set("user_id", user.ID)
				c.Set("username", user.Username)
				c.Set("role", user.Role)
				c.Next()
				return
			}
		}
		
		// 注意：不再允许 API_TOKEN 作为“超级管理员通行证”绕过所有权限。
		// 如果确实需要系统级token，应在重写版中限定到特定管理端点，并进行审计与RBAC控制。
		
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "无效的认证令牌",
		})
		c.Abort()
	}
}

