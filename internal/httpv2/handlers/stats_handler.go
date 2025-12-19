/**
 * v2 Stats Handler
 * - GET /api/v2/stats - 基础统计
 * - GET /api/v2/stats/aggregated - 聚合统计（日/周/月、来源、UA 等）
 */
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"short-link/internal/repo"
	"short-link/internal/service"
	"short-link/models"

	"github.com/gin-gonic/gin"
)

// StatsHandler 统计处理器（v2）
type StatsHandler struct {
	linkService *service.LinkService
	statsRepo   *repo.StatsRepo
	linkRepo    *repo.LinkRepo
}

// NewStatsHandler 创建 StatsHandler
func NewStatsHandler(linkService *service.LinkService, statsRepo *repo.StatsRepo, linkRepo *repo.LinkRepo) *StatsHandler {
	return &StatsHandler{
		linkService: linkService,
		statsRepo:   statsRepo,
		linkRepo:    linkRepo,
	}
}

// GetStats 获取基础统计信息
func (h *StatsHandler) GetStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	stats, err := h.linkService.GetStats(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计信息失败: " + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetAggregatedStats 获取聚合统计信息（日/周/月、来源、UA 等维度）
func (h *StatsHandler) GetAggregatedStats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	stats := &models.AggregatedStats{}

	// 基础统计
	linkStats, err := h.linkService.GetStats(ctx)
	if err == nil {
		stats.TotalLinks = linkStats.TotalLinks
		stats.TotalClicks = linkStats.TotalClicks
		stats.TodayClicks = linkStats.TodayClicks
		stats.TopLinks = linkStats.TopLinks
	}

	// 日统计（最近30天）
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if dailyStats, err := h.statsRepo.GetDailyStats(ctx, days); err == nil {
		stats.DailyStats = dailyStats
	}

	// 周统计（最近12周）
	weeks, _ := strconv.Atoi(c.DefaultQuery("weeks", "12"))
	if weeklyStats, err := h.statsRepo.GetWeeklyStats(ctx, weeks); err == nil {
		stats.WeeklyStats = weeklyStats
	}

	// 月统计（最近12个月）
	months, _ := strconv.Atoi(c.DefaultQuery("months", "12"))
	if monthlyStats, err := h.statsRepo.GetMonthlyStats(ctx, months); err == nil {
		stats.MonthlyStats = monthlyStats
	}

	// Top 来源
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if topReferers, err := h.statsRepo.GetTopReferers(ctx, limit); err == nil {
		stats.TopReferers = topReferers
	}

	// Top UA
	if topUAs, err := h.statsRepo.GetTopUserAgents(ctx, limit); err == nil {
		stats.TopUserAgents = topUAs
	}

	// Top IPs
	if topIPs, err := h.statsRepo.GetTopIPs(ctx, limit); err == nil {
		stats.TopIPs = topIPs
	}

	c.JSON(http.StatusOK, stats)
}


