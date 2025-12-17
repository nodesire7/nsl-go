/**
 * 数据库连接和初始化
 * 负责PostgreSQL数据库的连接、迁移和基础操作
 */
package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"
	"short-link/config"
	"short-link/models"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB 初始化数据库连接
func InitDB() error {
	dsn := config.AppConfig.GetDSN()
	
	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 连接池参数（sql.DB 自带连接池）
	if config.AppConfig.DBMaxOpenConns > 0 {
		DB.SetMaxOpenConns(config.AppConfig.DBMaxOpenConns)
	}
	if config.AppConfig.DBMaxIdleConns > 0 {
		DB.SetMaxIdleConns(config.AppConfig.DBMaxIdleConns)
	}
	if config.AppConfig.DBConnMaxLifetimeMinutes > 0 {
		DB.SetConnMaxLifetime(time.Duration(config.AppConfig.DBConnMaxLifetimeMinutes) * time.Minute)
	}

	// 测试连接
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接测试失败: %w", err)
	}

	log.Println("数据库连接成功")

	// 执行数据库迁移
	if err = Migrate(); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}

	// 安全回填：将历史明文 api_token 转为 hash，并清空明文字段
	if err = BackfillUserTokenHashes(); err != nil {
		return fmt.Errorf("token hash 回填失败: %w", err)
	}

	return nil
}

// Migrate 执行数据库迁移
func Migrate() error {
	queries := []string{
		// 创建用户表
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			api_token VARCHAR(255), -- 安全基线：不再明文存储/不再要求唯一与非空（仅保留 hash）
			role VARCHAR(20) DEFAULT 'user',
			max_links INTEGER DEFAULT 10,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// 用户索引
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE INDEX IF NOT EXISTS idx_users_api_token ON users(api_token)`,
		// 为重写版预留：token hash（逐步迁移）
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS api_token_hash VARCHAR(64)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_users_api_token_hash_unique ON users(api_token_hash) WHERE api_token_hash IS NOT NULL`,
		// 安全基线：不再要求明文 token（允许 NULL，移除唯一约束，后续将停止写入明文）
		`ALTER TABLE users ALTER COLUMN api_token DROP NOT NULL`,
		`ALTER TABLE users DROP CONSTRAINT IF EXISTS users_api_token_key`,
		`DROP INDEX IF EXISTS idx_users_api_token`,
		
		// 创建域名表
		`CREATE TABLE IF NOT EXISTS domains (
			id SERIAL PRIMARY KEY,
			user_id BIGINT DEFAULT 0,
			domain VARCHAR(255) NOT NULL,
			is_default BOOLEAN DEFAULT false,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, domain)
		)`,
		// 域名索引
		`CREATE INDEX IF NOT EXISTS idx_domains_user_id ON domains(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_domains_domain ON domains(domain)`,
		
		// 创建链接表（更新以支持用户和域名）
		`CREATE TABLE IF NOT EXISTS links (
			id SERIAL PRIMARY KEY,
			user_id BIGINT DEFAULT 0,
			domain_id BIGINT DEFAULT 0,
			code VARCHAR(255) NOT NULL,
			original_url TEXT NOT NULL,
			title VARCHAR(500),
			hash VARCHAR(64) NOT NULL,
			qr_code TEXT,
			click_count BIGINT DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(domain_id, code)
		)`,
		// 链接索引
		`CREATE INDEX IF NOT EXISTS idx_links_code ON links(code)`,
		`CREATE INDEX IF NOT EXISTS idx_links_hash ON links(hash)`,
		`CREATE INDEX IF NOT EXISTS idx_links_user_id ON links(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_links_domain_id ON links(domain_id)`,
		`CREATE INDEX IF NOT EXISTS idx_links_created_at ON links(created_at DESC)`,
		
		// 创建访问日志表
		`CREATE TABLE IF NOT EXISTS access_logs (
			id SERIAL PRIMARY KEY,
			link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
			ip VARCHAR(45),
			user_agent TEXT,
			referer TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		// 访问日志索引
		`CREATE INDEX IF NOT EXISTS idx_access_logs_link_id ON access_logs(link_id)`,
		`CREATE INDEX IF NOT EXISTS idx_access_logs_created_at ON access_logs(created_at DESC)`,
		
		// 创建配置表（用于存储管理员配置）
		`CREATE TABLE IF NOT EXISTS settings (
			key VARCHAR(255) PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		
		// 创建系统默认域名（如果不存在）
		`INSERT INTO domains (user_id, domain, is_default, is_active) 
		 SELECT 0, '', true, true 
		 WHERE NOT EXISTS (SELECT 1 FROM domains WHERE user_id = 0 AND is_default = true)`,
	}

	for _, query := range queries {
		if _, err := DB.Exec(query); err != nil {
			return fmt.Errorf("执行迁移失败: %w, SQL: %s", err, query)
		}
	}

	log.Println("数据库迁移完成")
	return nil
}

// InitAdminUser 初始化admin用户
func InitAdminUser() error {
	// 检查admin用户是否存在
	exists, err := CheckUsernameExists("admin")
	if err != nil {
		return fmt.Errorf("检查admin用户失败: %w", err)
	}
	
	if exists {
		return nil // admin用户已存在
	}
	
	// 生成随机密码
	randomPassword := generateRandomPassword(16)
	
	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(randomPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}
	
	// 生成API Token
	apiTokenBytes := make([]byte, 32)
	if _, err := rand.Read(apiTokenBytes); err != nil {
		return fmt.Errorf("生成API Token失败: %w", err)
	}
	apiToken := "nsl_" + hex.EncodeToString(apiTokenBytes)
	
	// 创建admin用户
	adminUser := &models.User{
		Username:  "admin",
		Email:      "admin@localhost",
		Password:   string(hashedPassword),
		APIToken:   apiToken,
		Role:       "admin",
		MaxLinks:   -1, // admin无限制
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	if err := CreateUser(adminUser); err != nil {
		return fmt.Errorf("创建admin用户失败: %w", err)
	}
	
	// 安全提示：默认不在日志中输出明文密码/token，避免日志泄漏导致全站失守
	log.Println("✅ Admin用户已创建（出于安全原因，不在日志中输出明文密码/Token）")
	log.Println("👉 请使用管理工具重置/查看：admin密码可通过 cmd/admin 工具重置")
	
	return nil
}

// generateRandomPassword 生成随机密码
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// 如果随机数生成失败，使用时间戳作为后备
		return fmt.Sprintf("admin%d", time.Now().Unix())
	}
	for i, b := range bytes {
		bytes[i] = charset[b%byte(len(charset))]
	}
	return string(bytes)
}

// CloseDB 关闭数据库连接
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

