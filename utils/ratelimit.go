/**
 * 限流工具
 * 实现滑动窗口限流算法（基于 Redis）
 * 实现 redo.md 2.5：限流策略优化（滑动窗口/令牌桶）
 */
package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SlidingWindowLimiter 滑动窗口限流器
type SlidingWindowLimiter struct {
	client  *redis.Client
	ctx     context.Context
	window  time.Duration // 时间窗口
	limit   int           // 窗口内允许的最大请求数
}

// NewSlidingWindowLimiter 创建滑动窗口限流器
func NewSlidingWindowLimiter(client *redis.Client, window time.Duration, limit int) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		client: client,
		ctx:    context.Background(),
		window: window,
		limit:  limit,
	}
}

// Allow 检查是否允许请求（滑动窗口算法）
// 使用 Redis Sorted Set 实现滑动窗口
func (l *SlidingWindowLimiter) Allow(key string) (bool, error) {
	if l.client == nil {
		// Redis 不可用时，允许请求（降级）
		return true, nil
	}

	now := time.Now()
	windowStart := now.Add(-l.window)

	// 1. 移除窗口外的旧记录
	pipe := l.client.Pipeline()
	pipe.ZRemRangeByScore(l.ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))
	
	// 2. 统计当前窗口内的请求数
	pipe.ZCard(l.ctx, key)
	
	// 3. 添加当前请求的时间戳
	pipe.ZAdd(l.ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	
	// 4. 设置过期时间（窗口长度 + 1秒缓冲）
	pipe.Expire(l.ctx, key, l.window+time.Second)
	
	results, err := pipe.Exec(l.ctx)
	if err != nil {
		return true, err // 出错时允许请求（降级）
	}

	// 获取当前窗口内的请求数（在添加当前请求之前）
	var count int64
	if len(results) >= 2 {
		countCmd, ok := results[1].(*redis.IntCmd)
		if ok {
			count = countCmd.Val()
		}
	}

	// 检查是否超过限制
	if count >= int64(l.limit) {
		return false, nil
	}

	return true, nil
}

// TokenBucketLimiter 令牌桶限流器（基于 Redis）
type TokenBucketLimiter struct {
	client    *redis.Client
	ctx       context.Context
	capacity  int           // 桶容量
	rate      float64       // 每秒生成的令牌数
	refillKey string        // 上次补充时间戳的 key
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(client *redis.Client, capacity int, rate float64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		client:    client,
		ctx:       context.Background(),
		capacity:  capacity,
		rate:      rate,
		refillKey: "token_bucket_refill",
	}
}

// Allow 检查是否允许请求（令牌桶算法）
func (l *TokenBucketLimiter) Allow(key string) (bool, error) {
	if l.client == nil {
		// Redis 不可用时，允许请求（降级）
		return true, nil
	}

	now := time.Now()
	tokensKey := key + ":tokens"
	lastRefillKey := key + ":last_refill"

	// 使用 Lua 脚本保证原子性
	luaScript := `
		local tokens_key = KEYS[1]
		local last_refill_key = KEYS[2]
		local capacity = tonumber(ARGV[1])
		local rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		
		local tokens = tonumber(redis.call('GET', tokens_key) or capacity)
		local last_refill = tonumber(redis.call('GET', last_refill_key) or now)
		
		-- 计算需要补充的令牌数
		local elapsed = now - last_refill
		local tokens_to_add = math.floor(elapsed * rate / 1000000000) -- 纳秒转秒
		
		if tokens_to_add > 0 then
			tokens = math.min(capacity, tokens + tokens_to_add)
			redis.call('SET', last_refill_key, now)
		end
		
		-- 检查是否有足够的令牌
		if tokens >= 1 then
			tokens = tokens - 1
			redis.call('SET', tokens_key, tokens)
			redis.call('EXPIRE', tokens_key, 3600)
			redis.call('EXPIRE', last_refill_key, 3600)
			return 1
		else
			redis.call('SET', tokens_key, tokens)
			redis.call('EXPIRE', tokens_key, 3600)
			redis.call('EXPIRE', last_refill_key, 3600)
			return 0
		end
	`

	result, err := l.client.Eval(l.ctx, luaScript, []string{tokensKey, lastRefillKey},
		l.capacity, l.rate, now.UnixNano()).Result()
	if err != nil {
		return true, err // 出错时允许请求（降级）
	}

	allowed, ok := result.(int64)
	if !ok {
		return true, nil
	}

	return allowed == 1, nil
}

