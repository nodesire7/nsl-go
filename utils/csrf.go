/**
 * CSRF 工具
 * 提供 CSRF Token 生成与校验（双提交 Cookie 方案）
 */
package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateCSRFToken 生成随机CSRF Token（32字节）
func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}


