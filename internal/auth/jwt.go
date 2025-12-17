/**
 * JWT 工具（重写版）
 * 负责签发与解析 JWT（用于 Cookie 登录态）
 */
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT Claims
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT 生成 JWT
func GenerateJWT(jwtSecret string, userID int64, username string, role string, ttl time.Duration) (string, error) {
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET 不能为空")
	}
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// ParseJWT 解析 JWT
func ParseJWT(jwtSecret string, tokenString string) (*Claims, error) {
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET 不能为空")
	}
	if tokenString == "" {
		return nil, errors.New("token 为空")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("不支持的JWT签名算法")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("无效token")
	}
	return claims, nil
}


