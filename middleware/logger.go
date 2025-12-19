/**
 * 日志中间件
 * 记录HTTP请求日志
 */
package middleware

import (
	"short-link/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 请求日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		
		c.Next()
		
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		
		utils.LogInfo("[%s] %s %s %d %v %s",
			method,
			path,
			clientIP,
			status,
			latency,
			c.Request.UserAgent(),
		)
	}
}

