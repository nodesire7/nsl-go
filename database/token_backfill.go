/**
 * Token Hash 回填
 * 将历史 users.api_token 明文计算 sha256 写入 users.api_token_hash，并清空明文列
 */
package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"
)

// BackfillUserTokenHashes 回填用户 token hash（不会输出任何明文 token）
func BackfillUserTokenHashes() error {
	if DB == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 选出需要回填的用户：hash 为空且明文存在
	rows, err := DB.QueryContext(ctx, `SELECT id, api_token FROM users WHERE (api_token_hash IS NULL OR api_token_hash = '') AND api_token IS NOT NULL AND api_token <> ''`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type row struct {
		id    int64
		token string
	}
	var items []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.token); err != nil {
			return err
		}
		items = append(items, r)
	}

	if len(items) == 0 {
		return nil
	}

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, it := range items {
		sum := sha256.Sum256([]byte(it.token))
		hash := hex.EncodeToString(sum[:])
		// 写入 hash 并清空明文 token
		if _, err := tx.ExecContext(ctx, `UPDATE users SET api_token_hash = $1, api_token = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, hash, it.id); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("已完成用户 token hash 回填：%d 条（明文 token 已清空）", len(items))
	return nil
}


