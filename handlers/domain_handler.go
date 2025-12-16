/**
 * 域名处理器
 * 处理域名相关的HTTP请求
 */
package handlers

import (
	"net/http"
	"short-link/models"
	"short-link/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// DomainHandler 域名处理器
type DomainHandler struct {
	domainService *services.DomainService
}

// NewDomainHandler 创建域名处理器实例
func NewDomainHandler(domainService *services.DomainService) *DomainHandler {
	return &DomainHandler{
		domainService: domainService,
	}
}

// CreateDomain 创建域名
func (h *DomainHandler) CreateDomain(c *gin.Context) {
	userID := c.GetInt64("user_id")
	
	var req models.CreateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}
	
	domain, err := h.domainService.CreateDomain(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, models.DomainResponse{
		ID:        domain.ID,
		Domain:    domain.Domain,
		IsDefault: domain.IsDefault,
		IsActive:  domain.IsActive,
		CreatedAt: domain.CreatedAt.Format("2006-01-02T15:04:05"),
	})
}

// GetDomains 获取用户的域名列表
func (h *DomainHandler) GetDomains(c *gin.Context) {
	userID := c.GetInt64("user_id")
	
	domains, err := h.domainService.GetUserDomains(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取域名列表失败: " + err.Error(),
		})
		return
	}
	
	var responses []models.DomainResponse
	for _, domain := range domains {
		responses = append(responses, models.DomainResponse{
			ID:        domain.ID,
			Domain:    domain.Domain,
			IsDefault: domain.IsDefault,
			IsActive:  domain.IsActive,
			CreatedAt: domain.CreatedAt.Format("2006-01-02T15:04:05"),
		})
	}
	
	c.JSON(http.StatusOK, responses)
}

// DeleteDomain 删除域名
func (h *DomainHandler) DeleteDomain(c *gin.Context) {
	userID := c.GetInt64("user_id")
	domainID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	
	if err := h.domainService.DeleteDomain(domainID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "域名已删除",
	})
}

// SetDefaultDomain 设置默认域名
func (h *DomainHandler) SetDefaultDomain(c *gin.Context) {
	userID := c.GetInt64("user_id")
	domainID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	
	if err := h.domainService.SetDefaultDomain(domainID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "默认域名已设置",
	})
}

