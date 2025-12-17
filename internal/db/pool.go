/**
 * 重写版数据库连接（pgxpool）
 * - 支持连接池参数
 * - 后续 repo/service 将基于该 pool 全量迁移
 */
package db

import (
	"context"
	"fmt"
	"time"

	appcfg "short-link/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool 包装 pgxpool.Pool
type Pool struct {
	*pgxpool.Pool
}

// New 创建连接池
func New(ctx context.Context, cfg *appcfg.Config) (*Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("解析DSN失败: %w", err)
	}

	poolCfg.MaxConns = cfg.DBMaxConns
	poolCfg.MaxConnLifetime = 30 * time.Minute
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	poolCfg.HealthCheckPeriod = 30 * time.Second

	p, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("创建连接池失败: %w", err)
	}

	// ping with timeout
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := p.Ping(pingCtx); err != nil {
		p.Close()
		return nil, fmt.Errorf("数据库Ping失败: %w", err)
	}

	return &Pool{Pool: p}, nil
}


