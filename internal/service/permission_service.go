/**
 * Permission Service（重写版）
 * - 权限检查业务逻辑
 */
package service

import (
	"context"
	"fmt"
	"short-link/internal/repo"
)

// PermissionService 权限服务
type PermissionService struct {
	permissionRepo *repo.PermissionRepo
}

// NewPermissionService 创建 PermissionService
func NewPermissionService(permissionRepo *repo.PermissionRepo) *PermissionService {
	return &PermissionService{permissionRepo: permissionRepo}
}

// CheckPermission 检查用户是否拥有指定权限
func (s *PermissionService) CheckPermission(ctx context.Context, userID int64, role string, permissionName string) (bool, error) {
	// admin 角色默认拥有所有权限（快速路径）
	if role == "admin" {
		return true, nil
	}
	return s.permissionRepo.CheckPermission(ctx, userID, role, permissionName)
}

// RequirePermission 要求用户拥有指定权限，否则返回错误
func (s *PermissionService) RequirePermission(ctx context.Context, userID int64, role string, permissionName string) error {
	hasPermission, err := s.CheckPermission(ctx, userID, role, permissionName)
	if err != nil {
		return fmt.Errorf("check permission failed: %w", err)
	}
	if !hasPermission {
		return fmt.Errorf("权限不足：需要 %s 权限", permissionName)
	}
	return nil
}

// GetUserPermissions 获取用户的所有权限
func (s *PermissionService) GetUserPermissions(ctx context.Context, userID int64, role string) ([]string, error) {
	// admin 角色默认拥有所有权限
	if role == "admin" {
		allPerms, err := s.permissionRepo.GetAllPermissions(ctx)
		if err != nil {
			return nil, err
		}
		permNames := make([]string, len(allPerms))
		for i, p := range allPerms {
			permNames[i] = p.Name
		}
		return permNames, nil
	}
	return s.permissionRepo.GetUserPermissions(ctx, userID, role)
}

