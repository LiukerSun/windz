package router

import (
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter(app *gin.Engine) {
	// 使用中间件
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.Cors())
	app.Use(middleware.RequestLogger())

	// API 路由组
	api := app.Group("/api/v1")

	// 注册认证路由
	registerUserRoutes(api)

	// 注册组织路由
	registerOrganizationRoutes(api)

}
