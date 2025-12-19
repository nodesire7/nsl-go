/**
 * Stats Repo（重写版）
 * - 负责聚合统计查询（日/周/月、来源、UA 等维度）
 * 实现 redo.md 5.3：聚合统计扩展
 */
package repo

import (
	"context"
	"fmt"
	"short-link/internal/db"
	"short-link/models"
	"time"
)

// StatsRepo 统计仓储
type StatsRepo struct {
	pool *db.Pool
}

// NewStatsRepo 创建 StatsRepo
func NewStatsRepo(pool *db.Pool) *StatsRepo {
	return &StatsRepo{pool: pool}
}

// GetDailyStats 获取日统计（最近30天）
func (r *StatsRepo) GetDailyStats(ctx context.Context, days int) ([]models.DailyStats, error) {
	if days <= 0 {
		days = 30
	}
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as click_count
		FROM access_logs
		WHERE created_at >= NOW() - INTERVAL '%d days'
		GROUP BY DATE(created_at)
		ORDER BY date DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, fmt.Sprintf(query, days), days)
	if err != nil {
		return nil, fmt.Errorf("get daily stats failed: %w", err)
	}
	defer rows.Close()

	var stats []models.DailyStats
	for rows.Next() {
		var s models.DailyStats
		if err := rows.Scan(&s.Date, &s.ClickCount); err != nil {
			return nil, fmt.Errorf("scan daily stats failed: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetWeeklyStats 获取周统计（最近12周）
func (r *StatsRepo) GetWeeklyStats(ctx context.Context, weeks int) ([]models.WeeklyStats, error) {
	if weeks <= 0 {
		weeks = 12
	}
	query := `
		SELECT 
			TO_CHAR(created_at, 'IYYY-"W"IW') as week,
			COUNT(*) as click_count
		FROM access_logs
		WHERE created_at >= NOW() - INTERVAL '%d weeks'
		GROUP BY TO_CHAR(created_at, 'IYYY-"W"IW')
		ORDER BY week DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, fmt.Sprintf(query, weeks), weeks)
	if err != nil {
		return nil, fmt.Errorf("get weekly stats failed: %w", err)
	}
	defer rows.Close()

	var stats []models.WeeklyStats
	for rows.Next() {
		var s models.WeeklyStats
		if err := rows.Scan(&s.Week, &s.ClickCount); err != nil {
			return nil, fmt.Errorf("scan weekly stats failed: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetMonthlyStats 获取月统计（最近12个月）
func (r *StatsRepo) GetMonthlyStats(ctx context.Context, months int) ([]models.MonthlyStats, error) {
	if months <= 0 {
		months = 12
	}
	query := `
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM') as month,
			COUNT(*) as click_count
		FROM access_logs
		WHERE created_at >= NOW() - INTERVAL '%d months'
		GROUP BY TO_CHAR(created_at, 'YYYY-MM')
		ORDER BY month DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, fmt.Sprintf(query, months), months)
	if err != nil {
		return nil, fmt.Errorf("get monthly stats failed: %w", err)
	}
	defer rows.Close()

	var stats []models.MonthlyStats
	for rows.Next() {
		var s models.MonthlyStats
		if err := rows.Scan(&s.Month, &s.ClickCount); err != nil {
			return nil, fmt.Errorf("scan monthly stats failed: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetTopReferers 获取 Top 来源（Top N）
func (r *StatsRepo) GetTopReferers(ctx context.Context, limit int) ([]models.RefererStats, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT 
			COALESCE(referer, 'direct') as referer,
			COUNT(*) as click_count
		FROM access_logs
		WHERE referer IS NOT NULL OR referer = ''
		GROUP BY COALESCE(referer, 'direct')
		ORDER BY click_count DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get top referers failed: %w", err)
	}
	defer rows.Close()

	var stats []models.RefererStats
	for rows.Next() {
		var s models.RefererStats
		if err := rows.Scan(&s.Referer, &s.ClickCount); err != nil {
			return nil, fmt.Errorf("scan top referers failed: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetTopUserAgents 获取 Top UA（Top N）
func (r *StatsRepo) GetTopUserAgents(ctx context.Context, limit int) ([]models.UserAgentStats, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT 
			COALESCE(user_agent, 'unknown') as user_agent,
			COUNT(*) as click_count
		FROM access_logs
		WHERE user_agent IS NOT NULL
		GROUP BY user_agent
		ORDER BY click_count DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get top user agents failed: %w", err)
	}
	defer rows.Close()

	var stats []models.UserAgentStats
	for rows.Next() {
		var s models.UserAgentStats
		if err := rows.Scan(&s.UserAgent, &s.ClickCount); err != nil {
			return nil, fmt.Errorf("scan top user agents failed: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetTopIPs 获取 Top IP（Top N）
func (r *StatsRepo) GetTopIPs(ctx context.Context, limit int) ([]models.IPStats, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `
		SELECT 
			COALESCE(ip, 'unknown') as ip,
			COUNT(*) as click_count
		FROM access_logs
		WHERE ip IS NOT NULL
		GROUP BY ip
		ORDER BY click_count DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get top ips failed: %w", err)
	}
	defer rows.Close()

	var stats []models.IPStats
	for rows.Next() {
		var s models.IPStats
		if err := rows.Scan(&s.IP, &s.ClickCount); err != nil {
			return nil, fmt.Errorf("scan top ips failed: %w", err)
		}
		stats = append(stats, s)
	}
	return stats, nil
}

// GetTodayClicks 获取今日点击数
func (r *StatsRepo) GetTodayClicks(ctx context.Context) (int64, error) {
	today := time.Now().Format("2006-01-02")
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM access_logs WHERE DATE(created_at) = $1`, today).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get today clicks failed: %w", err)
	}
	return count, nil
}

// GetTotalClicks 获取总点击数
func (r *StatsRepo) GetTotalClicks(ctx context.Context) (int64, error) {
	var count int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM access_logs`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get total clicks failed: %w", err)
	}
	return count, nil
}

