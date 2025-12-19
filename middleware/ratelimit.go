/**
 * 限流中间件
 * 提供API限流功能
 */
package middleware

import (
	"net/http"
	"short-link/cache"
	"short-link/utils"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var limiter = rate.NewLimiter(rate.Every(time.Second), 100) // 每秒100个请求

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果Redis可用，使用分布式限流
		if cache.RedisClient != nil {
			realIP := utils.GetRealIP(c.Request)
			key := "rate_limit:" + realIP
			count, err := cache.RedisClient.Incr(cache.Ctx, key).Result()
			if err == nil {
				if count == 1 {
					cache.RedisClient.Expire(cache.Ctx, key, time.Second)
				}
				if count > 100 { // 每秒100个请求
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "请求过于频繁，请稍后再试",
					})
					c.Abort()
					return
				}
			}
		} else {
			// 使用本地限流
			if !limiter.Allow() {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": "请求过于频繁，请稍后再试",
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

