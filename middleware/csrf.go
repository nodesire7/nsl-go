/**
 * CSRF 中间件
 * 使用“双提交 Cookie”方案：
 * - 服务端在登录/注册成功后下发 csrf_token Cookie（非 HttpOnly）
 * - 前端在 POST/PUT/PATCH/DELETE 请求中带上 Header: X-CSRF-Token
 * - 中间件校验 Header 与 Cookie 一致
 *
 * 注意：如果使用 Authorization: Bearer（用户 API Token）调用接口，则不需要 CSRF。
 */
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSRFMiddleware CSRF 防护中间件（仅对 Cookie 鉴权请求生效）
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 仅保护会产生副作用的请求
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		// 如果显式使用 Authorization Bearer（例如 API Token），跳过 CSRF 校验
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		// Cookie 模式：要求 X-CSRF-Token 与 csrf_token Cookie 一致
		cookieToken, _ := c.Cookie("csrf_token")
		headerToken := c.GetHeader("X-CSRF-Token")
		if cookieToken == "" || headerToken == "" || cookieToken != headerToken {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF校验失败",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}


