/**
 * v2 Stats Handler
 * - GET /api/v2/stats
 */
package handlers

import (
	"context"
	"net/http"
	"time"

	"short-link/internal/service"

	"github.com/gin-gonic/gin"
)

// StatsHandler 统计处理器（v2）
type StatsHandler struct {
	linkService *service.LinkService
}

// NewStatsHandler 创建 StatsHandler
func NewStatsHandler(linkService *service.LinkService) *StatsHandler {
	return &StatsHandler{linkService: linkService}
}

// GetStats 获取统计信息
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


