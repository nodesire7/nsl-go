/**
 * 用户数据模型
 * 定义用户的数据结构
 */
package models

import (
	"time"
)

// User 用户模型
type User struct {
	ID          int64     `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	Email       string    `json:"email" db:"email"`
	Password    string    `json:"-" db:"password"` // 不返回给前端
	APIToken    string    `json:"api_token" db:"api_token"` // 用户API Token
	Role        string    `json:"role" db:"role"`   // admin, user
	MaxLinks    int       `json:"max_links" db:"max_links"` // 最大链接数，-1表示无限制
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string     `json:"token"`
	User  UserInfo   `json:"user"`
}

// UserInfo 用户信息（不包含密码）
type UserInfo struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	APIToken  string `json:"api_token"` // 用户API Token
	Role      string `json:"role"`
	MaxLinks  int    `json:"max_links"`
	CreatedAt string `json:"created_at"`
}

// UpdateTokenRequest 更新Token请求
type UpdateTokenRequest struct {
	// 空结构，更新时会生成新token
}

