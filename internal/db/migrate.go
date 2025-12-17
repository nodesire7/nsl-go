/**
 * 重写版迁移器（轻量实现）
 * - 使用 go:embed 打包 SQL 文件
 * - 维护 schema_migrations(version) 防止重复执行
 *
 * 说明：这是过渡实现，后续可替换为 golang-migrate 或 goose。
 */
package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrate 执行迁移（按版本号升序）
func Migrate(ctx context.Context, pool *Pool) error {
	// 确保 migrations 表存在
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("创建schema_migrations失败: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("读取migrations目录失败: %w", err)
	}

	type mig struct {
		version int64
		name    string
	}
	var ms []mig
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 1 {
			continue
		}
		v, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			continue
		}
		ms = append(ms, mig{version: v, name: name})
	}

	sort.Slice(ms, func(i, j int) bool { return ms[i].version < ms[j].version })

	for _, m := range ms {
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, m.version).Scan(&exists); err != nil {
			return fmt.Errorf("检查迁移版本失败(%d): %w", m.version, err)
		}
		if exists {
			continue
		}

		b, err := migrationsFS.ReadFile("migrations/" + m.name)
		if err != nil {
			return fmt.Errorf("读取迁移文件失败(%s): %w", m.name, err)
		}

		sqlText := string(b)
		migCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, execErr := pool.Exec(migCtx, sqlText)
		cancel()
		if execErr != nil {
			return fmt.Errorf("执行迁移失败(%s): %w", m.name, execErr)
		}

		if _, err := pool.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, m.version); err != nil {
			return fmt.Errorf("写入schema_migrations失败(%d): %w", m.version, err)
		}
	}
	return nil
}


