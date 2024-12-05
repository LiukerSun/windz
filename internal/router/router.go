package router

import (
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine) {
	// 使用中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.Cors())

	// API 路由组
	api := r.Group("/api/v1")

	// 注册各个模块的路由
	registerUserRoutes(api)
	registerOrganizationRoutes(api)
}
