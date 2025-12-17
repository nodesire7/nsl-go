/**
 * 重写版配置（internal/config）
 * 与 legacy 的 config/config.go 并行存在，后续将逐步迁移并替换 legacy。
 */
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config 应用配置（重写版）
type Config struct {
	// 基础
	BaseURL     string
	ServerPort  int
	ServerMode  string
	LogLevel    string
	JWTSecret   string
	ReadTimeout time.Duration
	WriteTimeout time.Duration

	// 短链 code 长度配置（env 默认值，DB settings 可覆盖）
	MinCodeLength int
	MaxCodeLength int

	// PostgreSQL
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBMaxConns int32

	// Redis / Meilisearch（后续重写会启用）
	RedisHost     string
	RedisPassword string
	MeiliHost     string
	MeiliKey      string
}

// Load 从环境变量加载配置（重写版）
func Load() (*Config, error) {
	cfg := &Config{
		BaseURL:      getenv("BASE_URL", "http://localhost:9110"),
		ServerPort:   getenvInt("SERVER_PORT", 9110),
		ServerMode:   getenv("SERVER_MODE", "release"),
		LogLevel:     getenv("LOG_LEVEL", "INFO"),
		JWTSecret:    getenv("JWT_SECRET", ""),
		ReadTimeout:  time.Second * time.Duration(getenvInt("READ_TIMEOUT_SECONDS", 10)),
		WriteTimeout: time.Second * time.Duration(getenvInt("WRITE_TIMEOUT_SECONDS", 10)),
		MinCodeLength: getenvInt("MIN_CODE_LENGTH", 6),
		MaxCodeLength: getenvInt("MAX_CODE_LENGTH", 10),

		DBHost:     getenv("DB_HOST", "localhost"),
		DBPort:     getenvInt("DB_PORT", 5432),
		DBUser:     getenv("DB_USER", "postgres"),
		DBPassword: getenv("DB_PASSWORD", "postgres"),
		DBName:     getenv("DB_NAME", "shortlink"),
		DBSSLMode:  getenv("DB_SSLMODE", "disable"),
		DBMaxConns: int32(getenvInt("DB_MAX_CONNS", 20)),

		RedisHost:     getenv("REDIS_HOST", ""),
		RedisPassword: getenv("REDIS_PASSWORD", ""),
		MeiliHost:     getenv("MEILI_HOST", "http://localhost:7700"),
		MeiliKey:      getenv("MEILI_KEY", ""),
	}

	// 强制安全基线：生产/默认都要求 JWT_SECRET
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET 未设置：请设置强随机密钥（建议 openssl rand -hex 32）")
	}
	if cfg.MinCodeLength <= 0 || cfg.MaxCodeLength <= 0 || cfg.MinCodeLength > cfg.MaxCodeLength {
		return nil, fmt.Errorf("MIN_CODE_LENGTH / MAX_CODE_LENGTH 配置无效")
	}
	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}


