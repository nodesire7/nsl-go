/**
 * 链接处理器
 * 处理短链接相关的HTTP请求
 */
package handlers

import (
	"net/http"
	"short-link/database"
	"short-link/models"
	"short-link/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LinkHandler 链接处理器
type LinkHandler struct {
	linkService   *services.LinkService
	searchService *services.SearchService
	userService   *services.UserService
	domainService *services.DomainService
}

// NewLinkHandler 创建链接处理器实例
func NewLinkHandler(linkService *services.LinkService, searchService *services.SearchService, userService *services.UserService, domainService *services.DomainService) *LinkHandler {
	return &LinkHandler{
		linkService:   linkService,
		searchService: searchService,
		userService:   userService,
		domainService: domainService,
	}
}

// CreateLink 创建短链接
func (h *LinkHandler) CreateLink(c *gin.Context) {
	userID := c.GetInt64("user_id")
	
	// 检查用户链接限制
	if err := h.userService.CheckUserLinkLimit(userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	var req models.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}
	
	link, shortURL, err := h.linkService.CreateLink(userID, &req, h.domainService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "创建链接失败: " + err.Error(),
		})
		return
	}
	
	// 索引到搜索服务
	if h.searchService != nil {
		if err := h.searchService.IndexLink(link); err != nil {
			// 记录错误但不影响响应
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"id":           link.ID,
		"code":         link.Code,
		"short_url":    shortURL,
		"original_url": link.OriginalURL,
		"title":        link.Title,
		"qr_code":      link.QRCode,
		"click_count":  link.ClickCount,
		"created_at":   link.CreatedAt.Format("2006-01-02T15:04:05"),
	})
}

// GetLinks 获取链接列表
func (h *LinkHandler) GetLinks(c *gin.Context) {
	userID := c.GetInt64("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	links, total, err := h.linkService.GetLinks(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取链接列表失败: " + err.Error(),
		})
		return
	}
	
	var responses []models.LinkResponse
	for _, link := range links {
		// 获取域名信息
		domain, _ := database.GetDomainByID(link.DomainID)
		shortURL := h.domainService.GetDomainURL(domain) + "/" + link.Code
		
		responses = append(responses, models.LinkResponse{
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
	
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	
	c.JSON(http.StatusOK, models.PaginatedLinksResponse{
		Links:      responses,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	})
}

// SearchLinks 搜索链接
func (h *LinkHandler) SearchLinks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "搜索关键词不能为空",
		})
		return
	}
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	
	if h.searchService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "搜索服务未启用",
		})
		return
	}
	
	result, err := h.searchService.SearchLinks(query, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "搜索失败: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// GetLinkInfo 获取链接详情
func (h *LinkHandler) GetLinkInfo(c *gin.Context) {
	userID := c.GetInt64("user_id")
	code := c.Param("code")
	
	// 先尝试获取用户的链接
	link, err := h.linkService.GetLinkByCodeAnyDomain(code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "链接不存在",
		})
		return
	}
	
	// 检查权限（只能查看自己的链接，除非是管理员）
	if link.UserID != userID && userID != 0 {
		role := c.GetString("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "无权限访问",
			})
			return
		}
	}
	
	// 获取域名信息
	domain, _ := database.GetDomainByID(link.DomainID)
	shortURL := h.domainService.GetDomainURL(domain) + "/" + link.Code
	
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

// DeleteLink 删除链接
func (h *LinkHandler) DeleteLink(c *gin.Context) {
	userID := c.GetInt64("user_id")
	code := c.Param("code")
	
	if err := h.linkService.DeleteLink(userID, code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "删除链接失败: " + err.Error(),
		})
		return
	}
	
	// 从搜索索引中删除
	if h.searchService != nil {
		link, _ := h.linkService.GetLinkByCodeAnyDomain(code)
		if link != nil {
			_ = h.searchService.DeleteLink(link.ID)
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "链接已删除",
	})
}

// RedirectLink 重定向链接
func (h *LinkHandler) RedirectLink(c *gin.Context) {
	code := c.Param("code")
	
	originalURL, err := h.linkService.RedirectLink(
		code,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		c.GetHeader("Referer"),
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "链接不存在",
		})
		return
	}
	
	c.Redirect(http.StatusFound, originalURL)
}

