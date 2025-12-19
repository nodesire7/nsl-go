/**
 * AuditLog Repo（重写版）
 * - 负责 audit_logs 表写入（pgxpool）
 */
package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"short-link/internal/db"
	"short-link/models"
)

// AuditLogRepo 审计日志仓储
type AuditLogRepo struct {
	pool *db.Pool
}

// NewAuditLogRepo 创建 AuditLogRepo
func NewAuditLogRepo(pool *db.Pool) *AuditLogRepo {
	return &AuditLogRepo{pool: pool}
}

// CreateAuditLog 写入审计日志
func (r *AuditLogRepo) CreateAuditLog(ctx context.Context, log *models.AuditLog) error {
	var detailsJSON []byte
	var err error
	if log.Details != nil {
		detailsJSON, err = json.Marshal(log.Details)
		if err != nil {
			return fmt.Errorf("marshal details failed: %w", err)
		}
	}

	query := `
		INSERT INTO audit_logs (user_id, username, action, resource_type, resource_id, ip, user_agent, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	err = r.pool.QueryRow(
		ctx,
		query,
		log.UserID,
		log.Username,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.IP,
		log.UserAgent,
		detailsJSON,
		log.CreatedAt,
	).Scan(&log.ID)
	if err != nil {
		return fmt.Errorf("create audit log failed: %w", err)
	}
	return nil
}

