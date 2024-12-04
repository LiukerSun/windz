package main

import (
	"backend/internal/router"
	"backend/pkg/config"
	"backend/pkg/database"
	"backend/pkg/logger"

	"github.com/gin-gonic/gin"
)

func init() {
	// 初始化配置
	config.Init()

	// 初始化日志
	logger.Init(
		config.GetString("log.level"),
		config.GetString("log.format"),
		config.GetString("log.output"),
	)
	defer logger.Sync()

	// 初始化数据库
	database.Init()
}

func main() {
	// 创建Gin实例
	app := gin.New()

	// 初始化路由
	router.InitRouter(app)

	// 启动服务器
	port := config.GetString("app.port")
	if port == "" {
		port = "8080"
	}

	logger.Info("服务器启动在端口: " + port)
	if err := app.Run(":" + port); err != nil {
		logger.Fatal("服务器启动失败: " + err.Error())
	}
}
