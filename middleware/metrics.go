/**
 * Metrics 中间件
 * 收集 HTTP 请求的 Prometheus 指标
 */
package middleware

import (
	"strconv"
	"time"

	"short-link/internal/metrics"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware Prometheus 指标收集中间件
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		c.Next()

		// 记录请求延迟
		duration := time.Since(start).Seconds()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

		// 记录请求总数
		status := strconv.Itoa(c.Writer.Status())
		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
	}
}

