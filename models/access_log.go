/**
 * 访问日志模型
 * 记录短链接的访问记录
 */
package models

import (
	"time"
)

// AccessLog 访问日志
type AccessLog struct {
	ID        int64     `json:"id" db:"id"`
	LinkID    int64     `json:"link_id" db:"link_id"`
	IP        string    `json:"ip" db:"ip"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	Referer   string    `json:"referer" db:"referer"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// AccessStats 访问统计
type AccessStats struct {
	Date      string `json:"date"`
	ClickCount int64  `json:"click_count"`
}

