/**
 * Redis缓存
 * 提供Redis缓存功能
 */
package cache

import (
	"context"
	"fmt"
	"log"
	icfg "short-link/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis() error {
	// 从配置读取Redis地址，如果没有配置则不启用
	cfg, err := icfg.Load()
	if err != nil {
		log.Printf("加载配置失败，Redis将不可用: %v", err)
		return nil
	}
	
	redisHost := cfg.RedisHost
	if redisHost == "" {
		log.Println("Redis未配置，缓存功能将不可用")
		return nil
	}
	
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	
	// 测试连接
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Printf("Redis连接失败: %v，缓存功能将不可用", err)
		RedisClient = nil
		return nil
	}
	
	log.Println("Redis连接成功")
	return nil
}

// Get 获取缓存
func Get(key string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("Redis未启用")
	}
	return RedisClient.Get(Ctx, key).Result()
}

// Set 设置缓存
func Set(key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return nil // 如果Redis未启用，静默失败
	}
	return RedisClient.Set(Ctx, key, value, expiration).Err()
}

// Delete 删除缓存
func Delete(key string) error {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.Del(Ctx, key).Err()
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

