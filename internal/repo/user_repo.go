/**
 * 用户 Repo（重写版）
 * 使用 pgxpool + context 实现用户相关 DB 操作
 */
package repo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"short-link/internal/db"
	"short-link/models"

	"github.com/jackc/pgx/v5"
)

// UserRepo 用户仓储
type UserRepo struct {
	pool *db.Pool
}

// NewUserRepo 创建 UserRepo
func NewUserRepo(pool *db.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// TokenHash 计算 token hash（SHA256 hex）
func TokenHash(token string) string {
	if token == "" {
		return ""
	}
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// CreateUser 创建用户
func (r *UserRepo) CreateUser(ctx context.Context, u *models.User) error {
	if u == nil {
		return errors.New("user 不能为空")
	}

	tokenHash := TokenHash(u.APIToken)
	query := `
		INSERT INTO users (username, email, password, api_token, api_token_hash, role, max_links, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		u.Username,
		u.Email,
		u.Password,
		nil, // 安全：不再写入明文 api_token，仅写入 hash
		tokenHash,
		u.Role,
		u.MaxLinks,
		u.CreatedAt,
		u.UpdatedAt,
	).Scan(&u.ID)
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	return nil
}

// CheckUsernameExists 检查用户名是否存在
func (r *UserRepo) CheckUsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check username exists failed: %w", err)
	}
	return exists, nil
}

// CheckEmailExists 检查邮箱是否存在
func (r *UserRepo) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check email exists failed: %w", err)
	}
	return exists, nil
}

// GetUserByUsername 根据用户名获取用户
func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, username, email, password, COALESCE(api_token, ''), role, max_links, created_at, updated_at FROM users WHERE username = $1`
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.APIToken,
		&u.Role,
		&u.MaxLinks,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username failed: %w", err)
	}
	return u, nil
}

// GetUserByID 根据ID获取用户
func (r *UserRepo) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, username, email, password, COALESCE(api_token, ''), role, max_links, created_at, updated_at FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.APIToken,
		&u.Role,
		&u.MaxLinks,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id failed: %w", err)
	}
	return u, nil
}

// GetUserByToken 根据 API Token 获取用户（优先 hash，兼容明文）
func (r *UserRepo) GetUserByToken(ctx context.Context, token string) (*models.User, error) {
	u := &models.User{}
	tokenHash := TokenHash(token)
	query := `
		SELECT id, username, email, password, COALESCE(api_token, ''), role, max_links, created_at, updated_at
		FROM users
		WHERE api_token_hash = $1 OR api_token = $2
		LIMIT 1
	`
	err := r.pool.QueryRow(ctx, query, tokenHash, token).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.Password,
		&u.APIToken,
		&u.Role,
		&u.MaxLinks,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by token failed: %w", err)
	}
	return u, nil
}

// UpdateUserToken 更新用户 token（兼容同时更新明文与 hash）
func (r *UserRepo) UpdateUserToken(ctx context.Context, userID int64, newToken string) error {
	tokenHash := TokenHash(newToken)
	query := `UPDATE users SET api_token = NULL, api_token_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	ct, err := r.pool.Exec(ctx, query, tokenHash, userID)
	if err != nil {
		return fmt.Errorf("update user token failed: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateUserPassword 更新用户密码
func (r *UserRepo) UpdateUserPassword(ctx context.Context, username string, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = CURRENT_TIMESTAMP WHERE username = $2`
	ct, err := r.pool.Exec(ctx, query, hashedPassword, username)
	if err != nil {
		return fmt.Errorf("update user password failed: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetAdminUser 获取admin用户
func (r *UserRepo) GetAdminUser(ctx context.Context) (*models.User, error) {
	return r.GetUserByUsername(ctx, "admin")
}


