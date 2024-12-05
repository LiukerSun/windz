package database

import (
	"backend/internal/model"
	"backend/pkg/config"
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

// Init 初始化数据库连接
func Init() error {
	var err error
	dbType := config.GetString("database.type")

	// 配置GORM日志
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	}

	// 根据配置决定是否启用数据库日志
	if config.GetBool("database.enable_log") {
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
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// 自动迁移数据库结构
	if err := DB.AutoMigrate(
		&model.User{},
		&model.Organization{},
	); err != nil {
		return fmt.Errorf("failed to auto migrate database: %v", err)
	}

	// 初始化超级管理员账号
	if err := initSuperAdmin(); err != nil {
		return fmt.Errorf("failed to init super admin: %v", err)
	}

	return nil
}

func connectMySQL(gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.GetString("database.mysql.username"),
		config.GetString("database.mysql.password"),
		config.GetString("database.mysql.host"),
		config.GetInt("database.mysql.port"),
		config.GetString("database.mysql.database"),
	)
	return gorm.Open(mysql.Open(dsn), gormConfig)
}

func connectPostgres(gormConfig *gorm.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.GetString("database.postgres.host"),
		config.GetString("database.postgres.username"),
		config.GetString("database.postgres.password"),
		config.GetString("database.postgres.dbname"),
		config.GetInt("database.postgres.port"),
	)
	return gorm.Open(postgres.Open(dsn), gormConfig)
}

func connectSQLite(gormConfig *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(config.GetString("database.sqlite.database")), gormConfig)
}

// initSuperAdmin 初始化超级管理员账号
func initSuperAdmin() error {
	var count int64
	if err := DB.Model(&model.User{}).Where("role = ?", "super_admin").Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check super admin existence: %w", err)
	}

	// 如果已经存在超级管理员，则不需要创建
	if count > 0 {
		return nil
	}

	// 创建系统组织
	org := &model.Organization{
		Code:        "system",
		Description: "System Organization",
	}
	if err := DB.Create(org).Error; err != nil {
		return fmt.Errorf("failed to create system organization: %w", err)
	}

	// 获取默认密码
	defaultPassword := config.GetString("app.default_password")
	if defaultPassword == "" {
		defaultPassword = "admin123" // 如果没有配置，使用默认值
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建超级管理员账号
	admin := &model.User{
		Username:       "admin",
		Password:       string(hashedPassword),
		Role:           "super_admin",
		OrganizationID: org.ID,
	}
	if err := DB.Create(admin).Error; err != nil {
		return fmt.Errorf("failed to create super admin: %w", err)
	}

	return nil
}
