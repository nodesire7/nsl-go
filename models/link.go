/**
 * 链接数据模型
 * 定义短链接的数据结构
 */
package models

import (
	"time"
)

// Link 短链接模型
type Link struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`       // 用户ID，0表示系统
	DomainID    int64     `json:"domain_id" db:"domain_id"`   // 域名ID
	Code        string    `json:"code" db:"code"`
	OriginalURL string    `json:"original_url" db:"original_url"`
	Title       string    `json:"title" db:"title"`
	Hash        string    `json:"hash" db:"hash"` // URL内容的哈希值，用于一致性检查
	QRCode      string    `json:"qr_code" db:"qr_code"` // 二维码Base64
	ClickCount  int64     `json:"click_count" db:"click_count"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// LinkStats 链接统计信息
type LinkStats struct {
	TotalLinks    int64 `json:"total_links"`
	TotalClicks   int64 `json:"total_clicks"`
	TodayClicks   int64 `json:"today_clicks"`
	TopLinks      []Link `json:"top_links"`
}

// CreateLinkRequest 创建链接请求
type CreateLinkRequest struct {
	URL      string `json:"url" binding:"required"`
	Title    string `json:"title"`
	Code     string `json:"code"`      // 可选的自定义代码
	DomainID int64  `json:"domain_id"` // 可选，使用指定域名
}

// LinkResponse 链接响应
type LinkResponse struct {
	ID          int64  `json:"id"`
	Code        string `json:"code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	Title       string `json:"title"`
	QRCode      string `json:"qr_code"` // 二维码Base64
	ClickCount  int64  `json:"click_count"`
	CreatedAt   string `json:"created_at"`
}

// PaginatedLinksResponse 分页链接响应
type PaginatedLinksResponse struct {
	Links      []LinkResponse `json:"links"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

