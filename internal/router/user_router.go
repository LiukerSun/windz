package router

import (
	"backend/internal/controller"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func registerUserRoutes(api *gin.RouterGroup) {
	auth := &controller.Auth{}

	// 公开路由
	api.POST("/users/login", auth.Login)
	api.POST("/users/register", auth.Register)

	// 需要认证的路由
	api.GET("/users/me", middleware.RequireAuth(), auth.GetCurrentUser)
}
