/**
 * 统计处理器
 * 处理统计数据相关的HTTP请求
 */
package handlers

import (
	"net/http"
	"short-link/services"

	"github.com/gin-gonic/gin"
)

// StatsHandler 统计处理器
type StatsHandler struct {
	linkService *services.LinkService
}

// NewStatsHandler 创建统计处理器实例
func NewStatsHandler(linkService *services.LinkService) *StatsHandler {
	return &StatsHandler{
		linkService: linkService,
	}
}

// GetStats 获取统计信息
func (h *StatsHandler) GetStats(c *gin.Context) {
	stats, err := h.linkService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计信息失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, stats)
}

