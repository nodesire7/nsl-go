/**
 * 限流中间件
 * 提供API限流功能（滑动窗口 + 令牌桶）
 * 实现 redo.md 2.5：限流策略优化
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

var (
	// 本地限流器（令牌桶，作为降级方案）
	localLimiter = rate.NewLimiter(rate.Every(time.Second), 100) // 每秒100个请求
	
	// Redis 滑动窗口限流器（1分钟窗口，100个请求）
	slidingWindowLimiter *utils.SlidingWindowLimiter
	
	// Redis 令牌桶限流器（容量100，每秒补充100个令牌）
	tokenBucketLimiter *utils.TokenBucketLimiter
)

// InitRateLimiters 初始化限流器（在应用启动时调用）
func InitRateLimiters() {
	if cache.RedisClient != nil {
		// 滑动窗口：1分钟窗口，100个请求
		slidingWindowLimiter = utils.NewSlidingWindowLimiter(cache.RedisClient, time.Minute, 100)
		// 令牌桶：容量100，每秒补充100个令牌
		tokenBucketLimiter = utils.NewTokenBucketLimiter(cache.RedisClient, 100, 100.0)
	}
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		realIP := utils.GetRealIP(c.Request)
		key := "rate_limit:" + realIP
		
		var allowed bool
		var err error
		
		// 优先使用滑动窗口（更精确）
		if slidingWindowLimiter != nil {
			allowed, err = slidingWindowLimiter.Allow(key)
			if err == nil {
				if !allowed {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "请求过于频繁，请稍后再试",
					})
					c.Abort()
					return
				}
				c.Next()
				return
			}
		}
		
		// 降级到令牌桶
		if tokenBucketLimiter != nil {
			allowed, err = tokenBucketLimiter.Allow(key)
			if err == nil {
				if !allowed {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error": "请求过于频繁，请稍后再试",
					})
					c.Abort()
					return
				}
				c.Next()
				return
			}
		}
		
		// 最终降级到本地限流
		if !localLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

