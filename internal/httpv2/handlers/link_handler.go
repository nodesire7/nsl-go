/**
 * v2 Link Handler（重写版）
 * - POST /api/v2/links 创建短链
 * - GET  /api/v2/links 获取当前用户短链列表（分页）
 */
package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	appcfg "short-link/internal/config"
	"short-link/internal/repo"
	"short-link/internal/service"
	"short-link/models"

	"github.com/gin-gonic/gin"
)

// LinkHandler v2 链接处理器
type LinkHandler struct {
	cfg         *appcfg.Config
	linkService *service.LinkService
	linkRepo    *repo.LinkRepo
	domainRepo  *repo.DomainRepo
}

// NewLinkHandler 创建 LinkHandler
func NewLinkHandler(cfg *appcfg.Config, linkService *service.LinkService, linkRepo *repo.LinkRepo, domainRepo *repo.DomainRepo) *LinkHandler {
	return &LinkHandler{cfg: cfg, linkService: linkService, linkRepo: linkRepo, domainRepo: domainRepo}
}

// CreateLink 创建短链接
func (h *LinkHandler) CreateLink(c *gin.Context) {
	userID := c.GetInt64("user_id")

	var req models.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()
	link, shortURL, err := h.linkService.CreateLink(ctx, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.LinkResponse{
		ID:          link.ID,
		Code:        link.Code,
		ShortURL:    shortURL,
		OriginalURL: link.OriginalURL,
		Title:       link.Title,
		QRCode:      link.QRCode,
		ClickCount:  link.ClickCount,
		CreatedAt:   link.CreatedAt.Format("2006-01-02T15:04:05"),
	})
}

// GetLinks 获取当前用户的链接列表（分页）
func (h *LinkHandler) GetLinks(c *gin.Context) {
	userID := c.GetInt64("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()

	links, total, err := h.linkRepo.GetUserLinks(ctx, userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取链接列表失败: " + err.Error()})
		return
	}

	// domain 缓存：避免 N 次重复查询
	domainCache := map[int64]*models.Domain{}
	buildDomain := func(domainID int64) *models.Domain {
		if d, ok := domainCache[domainID]; ok {
			return d
		}
		if domainID <= 0 {
			domainCache[domainID] = nil
			return nil
		}
		d, err := h.domainRepo.GetDomainByID(ctx, domainID)
		if err != nil {
			domainCache[domainID] = nil
			return nil
		}
		domainCache[domainID] = d
		return d
	}

	resp := make([]models.LinkResponse, 0, len(links))
	for _, l := range links {
		d := buildDomain(l.DomainID)
		shortURL := h.linkService.BuildShortURL(d, l.Code)
		resp = append(resp, models.LinkResponse{
			ID:          l.ID,
			Code:        l.Code,
			ShortURL:    shortURL,
			OriginalURL: l.OriginalURL,
			Title:       l.Title,
			QRCode:      l.QRCode,
			ClickCount:  l.ClickCount,
			CreatedAt:   l.CreatedAt.Format("2006-01-02T15:04:05"),
		})
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	c.JSON(http.StatusOK, models.PaginatedLinksResponse{
		Links:      resp,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}


