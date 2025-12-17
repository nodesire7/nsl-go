/**
 * Settings Repo（重写版）
 * - 读取/写入 settings 表
 * - 目前 v2 主要用于读取 min/max code length（管理员可通过 v1 settings 或未来 v2 管理接口写入）
 */
package repo

import (
	"context"
	"errors"
	"fmt"
	"short-link/internal/db"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// SettingsRepo 配置仓储
type SettingsRepo struct {
	pool *db.Pool
}

// NewSettingsRepo 创建 SettingsRepo
func NewSettingsRepo(pool *db.Pool) *SettingsRepo {
	return &SettingsRepo{pool: pool}
}

// GetSetting 获取配置值
func (r *SettingsRepo) GetSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx, `SELECT value FROM settings WHERE key = $1`, key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get setting failed: %w", err)
	}
	return value, nil
}

// GetMinCodeLength 获取最小 code 长度（settings）
func (r *SettingsRepo) GetMinCodeLength(ctx context.Context) (int, error) {
	v, err := r.GetSetting(ctx, "min_code_length")
	if err != nil {
		return 0, err
	}
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("无效的min_code_length: %w", err)
	}
	return n, nil
}

// GetMaxCodeLength 获取最大 code 长度（settings）
func (r *SettingsRepo) GetMaxCodeLength(ctx context.Context) (int, error) {
	v, err := r.GetSetting(ctx, "max_code_length")
	if err != nil {
		return 0, err
	}
	if v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("无效的max_code_length: %w", err)
	}
	return n, nil
}


