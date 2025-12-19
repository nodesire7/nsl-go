/**
 * Permission Repo（重写版）
 * - 负责权限点相关 DB 操作（pgxpool）
 */
package repo

import (
	"context"
	"errors"
	"fmt"
	"short-link/internal/db"
	"short-link/models"

	"github.com/jackc/pgx/v5"
)

// PermissionRepo 权限仓储
type PermissionRepo struct {
	pool *db.Pool
}

// NewPermissionRepo 创建 PermissionRepo
func NewPermissionRepo(pool *db.Pool) *PermissionRepo {
	return &PermissionRepo{pool: pool}
}

// GetUserPermissions 获取用户的所有权限（包括角色权限和用户特定权限）
func (r *PermissionRepo) GetUserPermissions(ctx context.Context, userID int64, role string) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM permissions p
		WHERE p.id IN (
			-- 角色权限
			SELECT rp.permission_id
			FROM role_permissions rp
			WHERE rp.role = $1
			UNION
			-- 用户特定权限
			SELECT up.permission_id
			FROM user_permissions up
			WHERE up.user_id = $2
		)
	`
	rows, err := r.pool.Query(ctx, query, role, userID)
	if err != nil {
		return nil, fmt.Errorf("get user permissions failed: %w", err)
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan permission failed: %w", err)
		}
		permissions = append(permissions, name)
	}
	return permissions, nil
}

// CheckPermission 检查用户是否拥有指定权限
func (r *PermissionRepo) CheckPermission(ctx context.Context, userID int64, role string, permissionName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM permissions p
			WHERE p.name = $3
			AND p.id IN (
				-- 角色权限
				SELECT rp.permission_id
				FROM role_permissions rp
				WHERE rp.role = $1
				UNION
				-- 用户特定权限
				SELECT up.permission_id
				FROM user_permissions up
				WHERE up.user_id = $2
			)
		)
	`
	var hasPermission bool
	err := r.pool.QueryRow(ctx, query, role, userID, permissionName).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("check permission failed: %w", err)
	}
	return hasPermission, nil
}

// GrantPermissionToUser 授予用户特定权限
func (r *PermissionRepo) GrantPermissionToUser(ctx context.Context, userID int64, permissionName string) error {
	// 先获取权限 ID
	var permissionID int64
	err := r.pool.QueryRow(ctx, `SELECT id FROM permissions WHERE name = $1`, permissionName).Scan(&permissionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("permission not found: %s", permissionName)
	}
	if err != nil {
		return fmt.Errorf("get permission id failed: %w", err)
	}

	// 插入用户权限
	query := `INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2) ON CONFLICT (user_id, permission_id) DO NOTHING`
	_, err = r.pool.Exec(ctx, query, userID, permissionID)
	if err != nil {
		return fmt.Errorf("grant permission failed: %w", err)
	}
	return nil
}

// RevokePermissionFromUser 撤销用户特定权限
func (r *PermissionRepo) RevokePermissionFromUser(ctx context.Context, userID int64, permissionName string) error {
	// 先获取权限 ID
	var permissionID int64
	err := r.pool.QueryRow(ctx, `SELECT id FROM permissions WHERE name = $1`, permissionName).Scan(&permissionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("permission not found: %s", permissionName)
	}
	if err != nil {
		return fmt.Errorf("get permission id failed: %w", err)
	}

	// 删除用户权限
	query := `DELETE FROM user_permissions WHERE user_id = $1 AND permission_id = $2`
	_, err = r.pool.Exec(ctx, query, userID, permissionID)
	if err != nil {
		return fmt.Errorf("revoke permission failed: %w", err)
	}
	return nil
}

// GetAllPermissions 获取所有权限点
func (r *PermissionRepo) GetAllPermissions(ctx context.Context) ([]models.Permission, error) {
	query := `SELECT id, name, description, resource_type, created_at FROM permissions ORDER BY resource_type, name`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all permissions failed: %w", err)
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.ResourceType, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan permission failed: %w", err)
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}

