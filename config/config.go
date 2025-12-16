/**
 * 配置文件
 * 负责加载和管理应用配置
 */
package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构
type Config struct {
	// API配置
	APIToken string
	BaseURL  string

	// 数据库配置
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Meilisearch配置
	MeiliHost string
	MeiliKey  string

	// Redis配置
	RedisHost     string
	RedisPassword string

	// 短链接配置
	MinCodeLength int
	MaxCodeLength int

	// 日志配置
	LogLevel string

	// 服务器配置
	ServerPort int
	ServerMode string
}

var AppConfig *Config

// LoadConfig 加载配置文件
func LoadConfig() *Config {
	// 尝试加载.env文件（如果存在）
	_ = godotenv.Load()

	config := &Config{
		APIToken:      getEnv("API_TOKEN", ""),
		BaseURL:       getEnv("BASE_URL", "http://localhost:9110"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnvAsInt("DB_PORT", 5432),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		DBName:         getEnv("DB_NAME", "shortlink"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		MeiliHost:      getEnv("MEILI_HOST", "http://localhost:7700"),
		MeiliKey:       getEnv("MEILI_KEY", ""),
		RedisHost:      getEnv("REDIS_HOST", ""),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		MinCodeLength:  getEnvAsInt("MIN_CODE_LENGTH", 6),
		MaxCodeLength:  getEnvAsInt("MAX_CODE_LENGTH", 10),
		LogLevel:       getEnv("LOG_LEVEL", "INFO"),
		ServerPort:     getEnvAsInt("SERVER_PORT", 9110),
		ServerMode:     getEnv("SERVER_MODE", "release"),
	}

	// 验证必需配置
	if config.APIToken == "" {
		log.Fatal("API_TOKEN 环境变量未设置")
	}

	AppConfig = config
	return config
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt 获取环境变量并转换为整数
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return "host=" + c.DBHost +
		" port=" + strconv.Itoa(c.DBPort) +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

