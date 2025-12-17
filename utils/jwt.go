/**
 * JWT工具
 * 提供JWT token生成和验证功能
 */
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"short-link/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

// Claims JWT声明
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// InitJWT 初始化JWT密钥
func InitJWT() {
	// 使用独立的 JWT_SECRET，避免与 API_TOKEN 耦合、避免硬编码默认密钥
	if config.AppConfig != nil && config.AppConfig.JWTSecret != "" {
		jwtSecret = []byte(config.AppConfig.JWTSecret)
		return
	}

	// 未配置则生成临时密钥（仅保证可用性；重启后旧JWT会失效）
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// 极端兜底：使用固定字符串，但强烈警告
		jwtSecret = []byte("unsafe-fallback-jwt-secret")
		log.Println("⚠️ JWT_SECRET 未配置且随机生成失败：正在使用不安全的兜底密钥，请尽快设置 JWT_SECRET")
		return
	}
	jwtSecret = []byte(hex.EncodeToString(bytes))
	log.Println("⚠️ JWT_SECRET 未配置：已生成临时JWT密钥（重启后JWT失效），生产环境请设置 JWT_SECRET")
}

// GenerateToken 生成JWT token
func GenerateToken(userID int64, username, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

