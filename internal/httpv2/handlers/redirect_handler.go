/**
 * v2 Redirect Handler（重写版）
 * - GET /:code
 * 使用 pgxpool 解析 code（按 Host 匹配 domain），并写入点击/访问日志
 */
package handlers

import (
	"context"
	"net/http"
	"time"

	"short-link/internal/repo"
	"short-link/internal/service"

	"github.com/gin-gonic/gin"
)

// RedirectHandler v2 重定向处理器
type RedirectHandler struct {
	linkService *service.LinkService
}

// NewRedirectHandler 创建 RedirectHandler
func NewRedirectHandler(linkService *service.LinkService) *RedirectHandler {
	return &RedirectHandler{linkService: linkService}
}

// Redirect 执行 302 跳转
func (h *RedirectHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	url, err := h.linkService.RedirectLink(
		ctx,
		c.Request.Host,
		code,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		c.GetHeader("Referer"),
	)
	if err != nil {
		if err == repo.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "链接不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重定向失败: " + err.Error()})
		return
	}
	c.Redirect(http.StatusFound, url)
}


