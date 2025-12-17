/**
 * 安全响应头中间件
 * 提供基础的浏览器安全防护。
 */
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware 添加安全相关响应头
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 基础安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// CSP：当前页面含内联脚本，先允许 unsafe-inline（后续重写再收紧）
		c.Header("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; base-uri 'self'; frame-ancestors 'none'")

		// HSTS：仅在 HTTPS 时启用
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// 预检请求快速返回
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}


