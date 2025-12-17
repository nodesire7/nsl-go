/**
 * v2 鉴权中间件（重写版）
 * - Cookie: access_token（HttpOnly） => JWT
 * - Authorization: Bearer <token>
 *   - 若是 JWT：按 JWT 解析
 *   - 否则：按用户 API Token 解析（查库）
 */
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"short-link/internal/auth"
	"short-link/internal/repo"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 鉴权中间件
func AuthMiddleware(jwtSecret string, userRepo *repo.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// 1) Authorization Bearer 优先
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		}
		// 2) 否则尝试 Cookie access_token
		if token == "" {
			t, _ := c.Cookie("access_token")
			token = t
		}

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		// 尝试按 JWT 解析
		if claims, err := auth.ParseJWT(jwtSecret, token); err == nil {
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
			c.Next()
			return
		}

		// 回退：按 API Token 查库
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		u, err := userRepo.GetUserByToken(ctx, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效token"})
			c.Abort()
			return
		}
		c.Set("user_id", u.ID)
		c.Set("username", u.Username)
		c.Set("role", u.Role)
		c.Next()
	}
}


