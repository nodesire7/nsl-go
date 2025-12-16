/**
 * 域名数据模型
 * 定义短链接域名的数据结构
 */
package models

import (
	"time"
)

// Domain 域名模型
type Domain struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"` // 0表示系统默认域名
	Domain    string    `json:"domain" db:"domain"`   // 域名，如 example.com
	IsDefault bool      `json:"is_default" db:"is_default"` // 是否为默认域名
	IsActive  bool      `json:"is_active" db:"is_active"`    // 是否启用
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateDomainRequest 创建域名请求
type CreateDomainRequest struct {
	Domain    string `json:"domain" binding:"required"`
	IsDefault bool   `json:"is_default"`
}

// DomainResponse 域名响应
type DomainResponse struct {
	ID        int64  `json:"id"`
	Domain    string `json:"domain"`
	IsDefault bool   `json:"is_default"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

