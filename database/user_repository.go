/**
 * 用户数据访问层
 * 提供用户的数据库操作
 */
package database

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"short-link/models"
	"time"
)

// CreateUser 创建用户
func CreateUser(user *models.User) error {
	tokenHash := ""
	if user.APIToken != "" {
		h := sha256.Sum256([]byte(user.APIToken))
		tokenHash = hex.EncodeToString(h[:])
	}
	query := `
		INSERT INTO users (username, email, password, api_token, api_token_hash, role, max_links, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	
	err := DB.QueryRow(
		query,
		user.Username,
		user.Email,
		user.Password,
		nil, // 安全：不再写入明文 api_token，仅写入 hash
		tokenHash,
		user.Role,
		user.MaxLinks,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)
	
	return err
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password, COALESCE(api_token, ''), role, max_links, created_at, updated_at 
			  FROM users WHERE username = $1`
	
	err := DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.APIToken,
		&user.Role,
		&user.MaxLinks,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在")
	}
	
	return user, err
}

// GetUserByID 根据ID获取用户
func GetUserByID(userID int64) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, username, email, password, COALESCE(api_token, ''), role, max_links, created_at, updated_at 
			  FROM users WHERE id = $1`
	
	err := DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.APIToken,
		&user.Role,
		&user.MaxLinks,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("用户不存在")
	}
	
	return user, err
}

// GetUserByToken 根据Token获取用户
func GetUserByToken(token string) (*models.User, error) {
	user := &models.User{}
	h := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(h[:])
	// 兼容：优先用 api_token_hash 匹配（重写版），否则回退 api_token 明文（旧数据）
	query := `SELECT id, username, email, password, COALESCE(api_token, ''), role, max_links, created_at, updated_at 
			  FROM users 
			  WHERE api_token_hash = $1 OR api_token = $2
			  LIMIT 1`
	
	err := DB.QueryRow(query, tokenHash, token).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.APIToken,
		&user.Role,
		&user.MaxLinks,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("无效的token")
	}
	
	return user, err
}

// UpdateUserToken 更新用户Token
func UpdateUserToken(userID int64, newToken string) error {
	h := sha256.Sum256([]byte(newToken))
	tokenHash := hex.EncodeToString(h[:])
	// 安全：仅更新 hash，明文置空
	query := `UPDATE users SET api_token = NULL, api_token_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := DB.Exec(query, tokenHash, userID)
	return err
}

// UpdateUserTokenHash 更新用户Token Hash（重写版）
func UpdateUserTokenHash(userID int64, tokenHash string) error {
	query := `UPDATE users SET api_token_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := DB.Exec(query, tokenHash, userID)
	return err
}

// CheckUsernameExists 检查用户名是否存在
func CheckUsernameExists(username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	err := DB.QueryRow(query, username).Scan(&exists)
	return exists, err
}

// CheckEmailExists 检查邮箱是否存在
func CheckEmailExists(email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := DB.QueryRow(query, email).Scan(&exists)
	return exists, err
}

// GetUserLinkCount 获取用户的链接数量
func GetUserLinkCount(userID int64) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM links WHERE user_id = $1`
	err := DB.QueryRow(query, userID).Scan(&count)
	return count, err
}

// UpdateUserMaxLinks 更新用户最大链接数
func UpdateUserMaxLinks(userID int64, maxLinks int) error {
	query := `UPDATE users SET max_links = $1, updated_at = $2 WHERE id = $3`
	_, err := DB.Exec(query, maxLinks, time.Now(), userID)
	return err
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(username string, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE username = $2`
	result, err := DB.Exec(query, hashedPassword, username)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}
	
	return nil
}

// GetAdminUser 获取admin用户
func GetAdminUser() (*models.User, error) {
	return GetUserByUsername("admin")
}

