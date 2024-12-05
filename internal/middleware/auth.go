package middleware

import (
	"backend/internal/model"
	"backend/pkg/database"
	"backend/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireAuth 验证用户是否已登录
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 token
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "需要授权头信息"})
			c.Abort()
			return
		}

		// 移除 Bearer 前缀
		token = strings.TrimPrefix(token, "Bearer ")

		// 解析 token
		claims, err := jwt.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌"})
			c.Abort()
			return
		}

		// 从数据库获取用户信息
		var user model.User
		if err := database.DB.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("currentUser", &user)
		c.Next()
	}
}

// RequireSuperAdmin 验证用户是否为超级管理员
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户
		user, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			c.Abort()
			return
		}

		// 检查用户角色
		currentUser := user.(*model.User)
		if currentUser.Role != model.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要超级管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOrgAdmin 验证用户是否为组织管理员或超级管理员
func RequireOrgAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户
		user, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权"})
			c.Abort()
			return
		}

		// 检查用户角色
		currentUser := user.(*model.User)
		if currentUser.Role != model.RoleSuperAdmin && currentUser.Role != model.RoleOrgAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要组织管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}
