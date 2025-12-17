/**
 * Request-ID 中间件
 * 为每个请求生成/透传 X-Request-Id，便于日志关联与排查问题。
 */
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware 生成或透传请求ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Header("X-Request-Id", rid)
		c.Set("request_id", rid)
		c.Next()
	}
}


