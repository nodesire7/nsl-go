/**
 * 访问日志数据访问层
 * 提供访问日志的数据库操作
 */
package database

import (
	"fmt"
	"short-link/models"
)

// CreateAccessLog 创建访问日志
func CreateAccessLog(log *models.AccessLog) error {
	query := `
		INSERT INTO access_logs (link_id, ip, user_agent, referer, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	
	err := DB.QueryRow(
		query,
		log.LinkID,
		log.IP,
		log.UserAgent,
		log.Referer,
		log.CreatedAt,
	).Scan(&log.ID)
	
	return err
}

// GetAccessStats 获取访问统计（按日期）
func GetAccessStats(days int) ([]models.AccessStats, error) {
	query := `
		SELECT 
			DATE(created_at)::text as date,
			COUNT(*) as click_count
		FROM access_logs
		WHERE created_at >= NOW() - INTERVAL '1 day' * $1
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`
	
	rows, err := DB.Query(query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var stats []models.AccessStats
	for rows.Next() {
		var stat models.AccessStats
		err := rows.Scan(&stat.Date, &stat.ClickCount)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	
	return stats, nil
}

