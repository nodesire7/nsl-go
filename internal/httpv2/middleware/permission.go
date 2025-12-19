/**
 * 权限检查中间件（v2）
 * 基于 RBAC 权限点进行细粒度权限控制
 */
package middleware

import (
	"net/http"
	"short-link/internal/service"

	"github.com/gin-gonic/gin"
)

// RequirePermission 要求指定权限的中间件
func RequirePermission(permissionService *service.PermissionService, permissionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("user_id")
		role := c.GetString("role")

		ctx := c.Request.Context()
		if err := permissionService.RequirePermission(ctx, userID, role, permissionName); err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

