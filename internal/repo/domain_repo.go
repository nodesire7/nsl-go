/**
 * 域名 Repo（重写版）
 * - 负责读取默认域名、按ID读取域名并做权限校验
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

// DomainRepo 域名仓储
type DomainRepo struct {
	pool *db.Pool
}

// NewDomainRepo 创建 DomainRepo
func NewDomainRepo(pool *db.Pool) *DomainRepo {
	return &DomainRepo{pool: pool}
}

// GetDomainByID 根据ID获取域名
func (r *DomainRepo) GetDomainByID(ctx context.Context, domainID int64) (*models.Domain, error) {
	d := &models.Domain{}
	query := `SELECT id, user_id, domain, is_default, is_active, created_at, updated_at FROM domains WHERE id = $1`
	err := r.pool.QueryRow(ctx, query, domainID).Scan(
		&d.ID,
		&d.UserID,
		&d.Domain,
		&d.IsDefault,
		&d.IsActive,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get domain by id failed: %w", err)
	}
	return d, nil
}

// GetDefaultDomain 获取默认域名（先用户默认，再系统默认）
func (r *DomainRepo) GetDefaultDomain(ctx context.Context, userID int64) (*models.Domain, error) {
	d := &models.Domain{}

	// 用户默认
	query := `
		SELECT id, user_id, domain, is_default, is_active, created_at, updated_at
		FROM domains
		WHERE user_id = $1 AND is_default = true AND is_active = true
		LIMIT 1
	`
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&d.ID,
		&d.UserID,
		&d.Domain,
		&d.IsDefault,
		&d.IsActive,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err == nil {
		return d, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("get default domain failed: %w", err)
	}

	// 系统默认
	query = `
		SELECT id, user_id, domain, is_default, is_active, created_at, updated_at
		FROM domains
		WHERE user_id = 0 AND is_default = true AND is_active = true
		LIMIT 1
	`
	err = r.pool.QueryRow(ctx, query).Scan(
		&d.ID,
		&d.UserID,
		&d.Domain,
		&d.IsDefault,
		&d.IsActive,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		// 没有默认域名，返回一个空域名（由上层用 BaseURL 兜底）
		return &models.Domain{ID: 0, UserID: 0, Domain: "", IsDefault: true, IsActive: true}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get system default domain failed: %w", err)
	}
	return d, nil
}


