/**
 * JWT工具
 * 提供JWT token生成和验证功能
 */
package utils

import (
	"errors"
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
	// 使用API_TOKEN作为JWT密钥，如果没有则生成一个
	if config.AppConfig.APIToken != "" {
		jwtSecret = []byte(config.AppConfig.APIToken)
	} else {
		jwtSecret = []byte("default-secret-key-change-in-production")
	}
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

