package router

import (
	"backend/internal/controller"
	"backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// registerOrganizationRoutes 注册组织相关路由
func registerOrganizationRoutes(api *gin.RouterGroup) {
	org := &controller.Organization{}

	// 组织相关路由组
	orgGroup := api.Group("/organizations")
	orgGroup.Use(middleware.RequireAuth(), middleware.RequireSuperAdmin())
	{
		orgGroup.POST("", org.Create)       // 创建组织
		orgGroup.GET("", org.List)          // 获取组织列表
		orgGroup.GET("/:id", org.Get)       // 获取单个组织
		orgGroup.PUT("/:id", org.Update)    // 更新组织
		orgGroup.DELETE("/:id", org.Delete) // 删除组织
	}
}
