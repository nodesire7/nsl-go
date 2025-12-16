/**
 * 配置处理器
 * 处理系统配置相关的HTTP请求
 */
package handlers

import (
	"net/http"
	"short-link/config"
	"short-link/database"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SettingsHandler 配置处理器
type SettingsHandler struct{}

// NewSettingsHandler 创建配置处理器实例
func NewSettingsHandler() *SettingsHandler {
	return &SettingsHandler{}
}

// GetSettings 获取配置
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	minLength, _ := database.GetMinCodeLength()
	maxLength, _ := database.GetMaxCodeLength()
	
	// 如果数据库中没有配置，使用默认值
	if minLength == 0 {
		minLength = config.AppConfig.MinCodeLength
	}
	if maxLength == 0 {
		maxLength = config.AppConfig.MaxCodeLength
	}
	
	c.JSON(http.StatusOK, gin.H{
		"min_code_length": minLength,
		"max_code_length": maxLength,
	})
}

// UpdateSettings 更新配置
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var req struct {
		MinCodeLength *int `json:"min_code_length"`
		MaxCodeLength *int `json:"max_code_length"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求参数: " + err.Error(),
		})
		return
	}
	
	// 验证范围
	if req.MinCodeLength != nil {
		if *req.MinCodeLength < 3 || *req.MinCodeLength > 20 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "最小代码长度必须在3-20之间",
			})
			return
		}
		if err := database.SetMinCodeLength(*req.MinCodeLength); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "更新配置失败: " + err.Error(),
			})
			return
		}
		config.AppConfig.MinCodeLength = *req.MinCodeLength
	}
	
	if req.MaxCodeLength != nil {
		if *req.MaxCodeLength < 3 || *req.MaxCodeLength > 20 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "最大代码长度必须在3-20之间",
			})
			return
		}
		if err := database.SetMaxCodeLength(*req.MaxCodeLength); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "更新配置失败: " + err.Error(),
			})
			return
		}
		config.AppConfig.MaxCodeLength = *req.MaxCodeLength
	}
	
	// 验证最小长度不能大于最大长度
	minLength := config.AppConfig.MinCodeLength
	maxLength := config.AppConfig.MaxCodeLength
	if minLength > maxLength {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "最小代码长度不能大于最大代码长度",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置更新成功",
		"settings": gin.H{
			"min_code_length": minLength,
			"max_code_length": maxLength,
		},
	})
}

