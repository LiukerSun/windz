package database

import (
	"backend/internal/model"
	"backend/pkg/config"
	"backend/pkg/logger"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func Init() {
	var err error
	dbType := config.GetString("database.default")

	// 配置GORM日志
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	}

	// 根据配置决定是否启用数据库日志
	if config.GetBool("log.database_log") {
		gormConfig.Logger = gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  gormlogger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		)
	} else {
		gormConfig.Logger = gormlogger.Default.LogMode(gormlogger.Silent)
	}

	switch dbType {
	case "mysql":
		DB, err = connectMySQL(gormConfig)
	case "postgres":
		DB, err = connectPostgres(gormConfig)
	case "sqlite":
		DB, err = connectSQLite(gormConfig)
	default:
		panic(fmt.Sprintf("不支持的数据库类型: %s", dbType))
	}

	if err != nil {
		logger.Error("连接数据库失败: " + err.Error())
		panic(err)
	}

	logger.Info("数据库连接成功")

	// 自动迁移
	err = DB.AutoMigrate(
		&model.User{},
		&model.Organization{},
	)
	if err != nil {
		logger.Error("数据库迁移失败: " + err.Error())
		panic(err)
	}
	logger.Info("数据库迁移成功")

	// 初始化超级管理员
	initSuperAdmin()
}

func connectMySQL(gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		config.GetString("database.mysql.username"),
		config.GetString("database.mysql.password"),
		config.GetString("database.mysql.host"),
		config.GetString("database.mysql.port"),
		config.GetString("database.mysql.dbname"),
		config.GetString("database.mysql.charset"),
	)
	return gorm.Open(mysql.Open(dsn), gormConfig)
}

func connectPostgres(gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.GetString("database.postgres.host"),
		config.GetString("database.postgres.username"),
		config.GetString("database.postgres.password"),
		config.GetString("database.postgres.dbname"),
		config.GetString("database.postgres.port"),
		config.GetString("database.postgres.sslmode"),
	)
	return gorm.Open(postgres.Open(dsn), gormConfig)
}

func connectSQLite(gormConfig *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(config.GetString("database.sqlite.path")), gormConfig)
}

// initSuperAdmin 初始化超级管理员账号
func initSuperAdmin() {
	// 检查是否已经存在超级管理员
	var count int64
	DB.Model(&model.User{}).Where("role = ?", model.RoleSuperAdmin).Count(&count)
	if count > 0 {
		return // 已存在超级管理员，不需要继续
	}

	// 使用事务处理所有初始化操作
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 创建系统组织
		systemOrg := model.Organization{
			Code:        "system",
			Description: "System Organization",
		}
		if err := tx.Create(&systemOrg).Error; err != nil {
			return fmt.Errorf("failed to create system organization: %w", err)
		}

		// 创建超级管理员
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		superAdmin := model.User{
			Username: "admin",
			Password: string(hashedPassword),
			Email:    "admin@system.com",
			Role:     model.RoleSuperAdmin,
		}

		if err := tx.Create(&superAdmin).Error; err != nil {
			return fmt.Errorf("failed to create super admin: %w", err)
		}

		return nil
	})

	if err != nil {
		logger.Error("Failed to initialize system: " + err.Error())
		panic(err)
	}

	logger.Info("System initialized successfully with super admin")
}
