package router

import (
	"backend/internal/controller"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(api *gin.RouterGroup) {
	auth := &controller.Auth{}

	// 公开路由
	api.POST("/users/login", auth.Login)       // 用户登录
	api.POST("/users/register", auth.Register) // 用户注册

	// 需要认证的路由
	api.GET("/users/me", middleware.RequireAuth(), auth.GetCurrentUser)        // 获取当前用户信息
	api.POST("/users/password", middleware.RequireAuth(), auth.ChangePassword) // 修改密码

	// 超级管理员路由
	api.POST("/admin/login", auth.AdminLogin)                                                             // 管理员登录
	api.POST("/admin/create", middleware.RequireAuth(), middleware.RequireSuperAdmin(), auth.CreateAdmin) // 创建新的超级管理员
}
