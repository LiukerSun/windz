package main

import (
	"backend/internal/router"
	"backend/pkg/config"
	"backend/pkg/database"
	"backend/pkg/logger"
	"fmt"

	_ "backend/docs" // 导入swagger文档

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Windz Backend API
// @version         1.0
// @description     This is the API documentation for Windz Backend.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func init() {
	// 初始化配置
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("初始化配置失败: %v", err))
	}

	// 初始化日志
	logger.Init(
		config.GetString("log.level"),
		config.GetString("log.format"),
		config.GetString("log.output"),
	)

	// 初始化数据库连接
	if err := database.Init(); err != nil {
		logger.Error(fmt.Sprintf("初始化数据库失败: %v", err))
		panic(err)
	}
}

func main() {
	defer logger.Sync()

	// 创建gin实例
	r := gin.Default()

	// 注册路由
	router.RegisterRoutes(r)

	// 添加swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 启动服务器
	port := config.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	logger.Info("服务器启动在端口: " + port)
	if err := r.Run(":" + port); err != nil {
		logger.Fatal("服务器启动失败: " + err.Error())
	}
}
