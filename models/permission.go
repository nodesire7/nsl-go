/**
 * 权限点模型
 * RBAC 权限点系统
 */
package models

import (
	"time"
)

// Permission 权限点
type Permission struct {
	ID           int64     `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// UserPermission 用户权限关联
type UserPermission struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	PermissionID int64    `json:"permission_id" db:"permission_id"`
	GrantedAt   time.Time `json:"granted_at" db:"granted_at"`
}

// RolePermission 角色权限关联
type RolePermission struct {
	ID           int64     `json:"id" db:"id"`
	Role         string    `json:"role" db:"role"`
	PermissionID int64     `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

