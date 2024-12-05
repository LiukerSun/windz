package router

import (
	"backend/internal/controller"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerUserRoutes 注册用户相关路由
func registerUserRoutes(api *gin.RouterGroup) {
	authController := controller.NewAuth()

	// 认证相关路由
	auth := api.Group("/auth")
	{
		auth.POST("/login", authController.Login)       // 用户登录
		auth.POST("/register", authController.Register) // 用户注册

		// 需要认证的路由
		authRequired := auth.Use(middleware.RequireAuth())
		{
			authRequired.POST("/change-password", authController.ChangePassword) // 修改密码
			authRequired.POST("/reset-password", authController.ResetPassword)   // 重置密码
			authRequired.POST("/create-admin", authController.CreateAdmin)       // 创建管理员
		}
	}
}
