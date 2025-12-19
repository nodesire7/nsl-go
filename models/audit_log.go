/**
 * 审计日志模型
 * 记录管理员操作和敏感操作
 */
package models

import (
	"time"
)

// AuditLog 审计日志
type AuditLog struct {
	ID           int64                  `json:"id" db:"id"`
	UserID       *int64                 `json:"user_id,omitempty" db:"user_id"`
	Username     string                 `json:"username,omitempty" db:"username"`
	Action       string                 `json:"action" db:"action"`
	ResourceType string                 `json:"resource_type,omitempty" db:"resource_type"`
	ResourceID   *int64                 `json:"resource_id,omitempty" db:"resource_id"`
	IP           string                 `json:"ip,omitempty" db:"ip"`
	UserAgent    string                 `json:"user_agent,omitempty" db:"user_agent"`
	Details      map[string]interface{} `json:"details,omitempty" db:"details"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

