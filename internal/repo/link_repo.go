/**
 * Link Repo（重写版）
 * - 负责 links 表的读写（pgxpool）
 * - 依赖数据库唯一约束 (domain_id, code) 实现并发安全
 */
package repo

import (
	"context"
	"errors"
	"fmt"
	"short-link/internal/db"
	"short-link/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// LinkRepo 链接仓储
type LinkRepo struct {
	pool *db.Pool
}

// NewLinkRepo 创建 LinkRepo
func NewLinkRepo(pool *db.Pool) *LinkRepo {
	return &LinkRepo{pool: pool}
}

// IsUniqueViolation 判断是否唯一约束冲突（23505）
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// CreateLink 创建链接
func (r *LinkRepo) CreateLink(ctx context.Context, link *models.Link) error {
	query := `
		INSERT INTO links (user_id, domain_id, code, original_url, title, hash, qr_code, click_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		link.UserID,
		link.DomainID,
		link.Code,
		link.OriginalURL,
		link.Title,
		link.Hash,
		link.QRCode,
		link.ClickCount,
		link.CreatedAt,
		link.UpdatedAt,
	).Scan(&link.ID)
	if err != nil {
		return fmt.Errorf("create link failed: %w", err)
	}
	return nil
}

// GetLinkByHashUserDomain 幂等检查：按 (hash, user_id, domain_id)
func (r *LinkRepo) GetLinkByHashUserDomain(ctx context.Context, hash string, userID int64, domainID int64) (*models.Link, error) {
	l := &models.Link{}
	query := `
		SELECT id, user_id, domain_id, code, original_url, title, hash, qr_code, click_count, created_at, updated_at
		FROM links
		WHERE hash = $1 AND user_id = $2 AND domain_id = $3
		LIMIT 1
	`
	err := r.pool.QueryRow(ctx, query, hash, userID, domainID).Scan(
		&l.ID,
		&l.UserID,
		&l.DomainID,
		&l.Code,
		&l.OriginalURL,
		&l.Title,
		&l.Hash,
		&l.QRCode,
		&l.ClickCount,
		&l.CreatedAt,
		&l.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get link by hash user domain failed: %w", err)
	}
	return l, nil
}

// CheckCodeExistsInDomain 检查 code 在 domain_id 下是否存在
func (r *LinkRepo) CheckCodeExistsInDomain(ctx context.Context, code string, domainID int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM links WHERE code = $1 AND domain_id = $2)`, code, domainID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check code exists failed: %w", err)
	}
	return exists, nil
}

// GetCodeCountByLength 获取指定长度 code 数量（全表）
func (r *LinkRepo) GetCodeCountByLength(ctx context.Context, length int) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM links WHERE LENGTH(code) = $1`, length).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get code count by length failed: %w", err)
	}
	return count, nil
}

// GetUserLinks 获取用户链接分页列表
func (r *LinkRepo) GetUserLinks(ctx context.Context, userID int64, page int, limit int) ([]models.Link, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM links WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count links failed: %w", err)
	}

	query := `
		SELECT id, user_id, domain_id, code, original_url, title, hash, qr_code, click_count, created_at, updated_at
		FROM links
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list links failed: %w", err)
	}
	defer rows.Close()

	var links []models.Link
	for rows.Next() {
		var l models.Link
		if err := rows.Scan(
			&l.ID,
			&l.UserID,
			&l.DomainID,
			&l.Code,
			&l.OriginalURL,
			&l.Title,
			&l.Hash,
			&l.QRCode,
			&l.ClickCount,
			&l.CreatedAt,
			&l.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan link failed: %w", err)
		}
		links = append(links, l)
	}
	return links, total, nil
}

// DeleteUserLink 删除用户在指定 domain 下的链接
func (r *LinkRepo) DeleteUserLink(ctx context.Context, userID int64, domainID int64, code string) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM links WHERE user_id = $1 AND domain_id = $2 AND code = $3`, userID, domainID, code)
	if err != nil {
		return fmt.Errorf("delete link failed: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// CountLinksByUser 统计用户链接数量（用于 max_links 限制）
func (r *LinkRepo) CountLinksByUser(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM links WHERE user_id = $1`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count links by user failed: %w", err)
	}
	return count, nil
}

// IncrementClickCount 增加点击计数（v2 预留）
func (r *LinkRepo) IncrementClickCount(ctx context.Context, linkID int64) error {
	_, err := r.pool.Exec(ctx, `UPDATE links SET click_count = click_count + 1, updated_at = $1 WHERE id = $2`, time.Now(), linkID)
	if err != nil {
		return fmt.Errorf("increment click_count failed: %w", err)
	}
	return nil
}


