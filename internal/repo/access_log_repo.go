/**
 * AccessLog Repo（重写版）
 * - 负责 access_logs 表写入（pgxpool）
 */
package repo

import (
	"context"
	"fmt"
	"short-link/internal/db"
	"short-link/models"
)

// AccessLogRepo 访问日志仓储
type AccessLogRepo struct {
	pool *db.Pool
}

// NewAccessLogRepo 创建 AccessLogRepo
func NewAccessLogRepo(pool *db.Pool) *AccessLogRepo {
	return &AccessLogRepo{pool: pool}
}

// CreateAccessLog 写入访问日志
func (r *AccessLogRepo) CreateAccessLog(ctx context.Context, log *models.AccessLog) error {
	query := `INSERT INTO access_logs (link_id, ip, user_agent, referer, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	if err := r.pool.QueryRow(ctx, query, log.LinkID, log.IP, log.UserAgent, log.Referer, log.CreatedAt).Scan(&log.ID); err != nil {
		return fmt.Errorf("create access log failed: %w", err)
	}
	return nil
}


