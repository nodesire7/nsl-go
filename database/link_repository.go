/**
 * 链接数据访问层
 * 提供链接的数据库操作
 */
package database

import (
	"database/sql"
	"fmt"
	"short-link/config"
	"short-link/models"
	"time"
)

// CreateLink 创建新链接
func CreateLink(link *models.Link) error {
	query := `
		INSERT INTO links (user_id, domain_id, code, original_url, title, hash, qr_code, click_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	
	err := DB.QueryRow(
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
	
	return err
}

// GetLinkByCode 根据代码获取链接（支持域名）
func GetLinkByCode(code string, domainID int64) (*models.Link, error) {
	link := &models.Link{}
	query := `SELECT id, user_id, domain_id, code, original_url, title, hash, qr_code, click_count, created_at, updated_at 
			  FROM links WHERE code = $1 AND domain_id = $2`
	
	err := DB.QueryRow(query, code, domainID).Scan(
		&link.ID,
		&link.UserID,
		&link.DomainID,
		&link.Code,
		&link.OriginalURL,
		&link.Title,
		&link.Hash,
		&link.QRCode,
		&link.ClickCount,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("链接不存在")
	}
	
	return link, err
}

// GetLinkByCodeAnyDomain 根据代码获取链接（任意域名，用于重定向）
func GetLinkByCodeAnyDomain(code string) (*models.Link, error) {
	link := &models.Link{}
	query := `SELECT id, user_id, domain_id, code, original_url, title, hash, qr_code, click_count, created_at, updated_at 
			  FROM links WHERE code = $1 LIMIT 1`
	
	err := DB.QueryRow(query, code).Scan(
		&link.ID,
		&link.UserID,
		&link.DomainID,
		&link.Code,
		&link.OriginalURL,
		&link.Title,
		&link.Hash,
		&link.QRCode,
		&link.ClickCount,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("链接不存在")
	}
	
	return link, err
}

// GetLinkByHash 根据哈希获取链接（用于一致性检查，已废弃，使用GetLinkByHashAndUser）
func GetLinkByHash(hash string) (*models.Link, error) {
	return GetLinkByHashAndUser(hash, 0)
}

// GetLinkByHashAndUser 根据哈希和用户ID获取链接（用于一致性检查）
func GetLinkByHashAndUser(hash string, userID int64) (*models.Link, error) {
	link := &models.Link{}
	query := `SELECT id, user_id, domain_id, code, original_url, title, hash, qr_code, click_count, created_at, updated_at 
			  FROM links WHERE hash = $1 AND user_id = $2 LIMIT 1`
	
	err := DB.QueryRow(query, hash, userID).Scan(
		&link.ID,
		&link.UserID,
		&link.DomainID,
		&link.Code,
		&link.OriginalURL,
		&link.Title,
		&link.Hash,
		&link.QRCode,
		&link.ClickCount,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return link, err
}

// CheckCodeExists 检查代码是否已存在（已废弃，使用CheckCodeExistsInDomain）
func CheckCodeExists(code string) (bool, error) {
	return CheckCodeExistsInDomain(code, 0)
}

// CheckCodeExistsInDomain 检查代码在指定域名下是否已存在
func CheckCodeExistsInDomain(code string, domainID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM links WHERE code = $1 AND domain_id = $2)`
	err := DB.QueryRow(query, code, domainID).Scan(&exists)
	return exists, err
}

// IncrementClickCount 增加点击计数
func IncrementClickCount(linkID int64) error {
	query := `UPDATE links SET click_count = click_count + 1, updated_at = $1 WHERE id = $2`
	_, err := DB.Exec(query, time.Now(), linkID)
	return err
}

// GetLinks 获取链接列表（分页，已废弃，使用GetUserLinks）
func GetLinks(page, limit int) ([]models.Link, int64, error) {
	return GetUserLinks(0, page, limit)
}

// GetUserLinks 获取用户的链接列表（分页）
func GetUserLinks(userID int64, page, limit int) ([]models.Link, int64, error) {
	offset := (page - 1) * limit
	
	// 获取总数
	var total int64
	var countQuery string
	if userID == 0 {
		countQuery = `SELECT COUNT(*) FROM links`
		err := DB.QueryRow(countQuery).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		countQuery = `SELECT COUNT(*) FROM links WHERE user_id = $1`
		err := DB.QueryRow(countQuery, userID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}
	
	// 获取链接列表
	var query string
	var rows *sql.Rows
	var err error
	if userID == 0 {
		query = `SELECT id, user_id, domain_id, code, original_url, title, hash, qr_code, click_count, created_at, updated_at 
				 FROM links ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		rows, err = DB.Query(query, limit, offset)
	} else {
		query = `SELECT id, user_id, domain_id, code, original_url, title, hash, qr_code, click_count, created_at, updated_at 
				 FROM links WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		rows, err = DB.Query(query, userID, limit, offset)
	}
	
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var links []models.Link
	for rows.Next() {
		var link models.Link
		err := rows.Scan(
			&link.ID,
			&link.UserID,
			&link.DomainID,
			&link.Code,
			&link.OriginalURL,
			&link.Title,
			&link.Hash,
			&link.QRCode,
			&link.ClickCount,
			&link.CreatedAt,
			&link.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		links = append(links, link)
	}
	
	return links, total, nil
}

// DeleteLink 删除链接（已废弃，使用DeleteUserLink）
func DeleteLink(code string) error {
	return DeleteUserLink(0, code)
}

// DeleteUserLink 删除用户的链接
func DeleteUserLink(userID int64, code string) error {
	var query string
	var result sql.Result
	var err error
	
	if userID == 0 {
		query = `DELETE FROM links WHERE code = $1`
		result, err = DB.Exec(query, code)
	} else {
		query = `DELETE FROM links WHERE code = $1 AND user_id = $2`
		result, err = DB.Exec(query, code, userID)
	}
	
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("链接不存在或无权限删除")
	}
	
	return nil
}

// GetLinkStats 获取链接统计信息
func GetLinkStats() (*models.LinkStats, error) {
	stats := &models.LinkStats{}
	
	// 总链接数
	err := DB.QueryRow(`SELECT COUNT(*) FROM links`).Scan(&stats.TotalLinks)
	if err != nil {
		return nil, err
	}
	
	// 总点击数
	err = DB.QueryRow(`SELECT COALESCE(SUM(click_count), 0) FROM links`).Scan(&stats.TotalClicks)
	if err != nil {
		return nil, err
	}
	
	// 今日点击数（通过访问日志）
	today := time.Now().Format("2006-01-02")
	err = DB.QueryRow(
		`SELECT COUNT(*) FROM access_logs WHERE DATE(created_at) = $1`,
		today,
	).Scan(&stats.TodayClicks)
	if err != nil {
		return nil, err
	}
	
	// 热门链接（前10）
	query := `SELECT id, code, original_url, title, hash, click_count, created_at, updated_at 
			  FROM links ORDER BY click_count DESC LIMIT 10`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var link models.Link
		err := rows.Scan(
			&link.ID,
			&link.Code,
			&link.OriginalURL,
			&link.Title,
			&link.Hash,
			&link.ClickCount,
			&link.CreatedAt,
			&link.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		stats.TopLinks = append(stats.TopLinks, link)
	}
	
	return stats, nil
}

// GetCodeCountByLength 获取指定长度的代码数量
func GetCodeCountByLength(length int) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM links WHERE LENGTH(code) = $1`
	err := DB.QueryRow(query, length).Scan(&count)
	return count, err
}

// GetMaxCodeLength 获取当前使用的最大代码长度
func GetMaxCodeLength() (int, error) {
	var maxLength sql.NullInt64
	query := `SELECT MAX(LENGTH(code)) FROM links`
	err := DB.QueryRow(query).Scan(&maxLength)
	if err != nil {
		return 0, err
	}
	if !maxLength.Valid {
		return config.AppConfig.MinCodeLength, nil
	}
	return int(maxLength.Int64), nil
}

