/**
 * 统计模型
 * 聚合统计扩展：日/周/月、来源、UA 等维度
 */
package models

// DailyStats 日统计
type DailyStats struct {
	Date       string `json:"date"`
	ClickCount int64  `json:"click_count"`
}

// WeeklyStats 周统计
type WeeklyStats struct {
	Week       string `json:"week"` // 格式：2025-W01
	ClickCount int64  `json:"click_count"`
}

// MonthlyStats 月统计
type MonthlyStats struct {
	Month      string `json:"month"` // 格式：2025-01
	ClickCount int64  `json:"click_count"`
}

// RefererStats 来源统计
type RefererStats struct {
	Referer    string `json:"referer"`
	ClickCount int64  `json:"click_count"`
}

// UserAgentStats UA 统计
type UserAgentStats struct {
	UserAgent  string `json:"user_agent"`
	ClickCount int64  `json:"click_count"`
}

// IPStats IP 统计
type IPStats struct {
	IP         string `json:"ip"`
	ClickCount int64  `json:"click_count"`
}

// AggregatedStats 聚合统计响应
type AggregatedStats struct {
	// 基础统计
	TotalLinks    int64 `json:"total_links"`
	TotalClicks   int64 `json:"total_clicks"`
	TodayClicks   int64 `json:"today_clicks"`
	
	// 时间维度统计
	DailyStats    []DailyStats    `json:"daily_stats,omitempty"`    // 最近30天
	WeeklyStats   []WeeklyStats   `json:"weekly_stats,omitempty"`   // 最近12周
	MonthlyStats  []MonthlyStats   `json:"monthly_stats,omitempty"`  // 最近12个月
	
	// 来源维度统计
	TopReferers   []RefererStats  `json:"top_referers,omitempty"`   // Top 10
	TopUserAgents []UserAgentStats `json:"top_user_agents,omitempty"` // Top 10
	TopIPs        []IPStats       `json:"top_ips,omitempty"`        // Top 10
	
	// 链接维度统计
	TopLinks      []Link          `json:"top_links,omitempty"`      // Top 10
}

