package config

import (
	"github.com/spf13/viper"
)

var Config *viper.Viper

// Init 初始化配置
func Init() error {
	Config = viper.New()
	Config.SetConfigName("config")
	Config.SetConfigType("yaml")
	Config.AddConfigPath("./config")

	if err := Config.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// GetString 获取字符串配置
func GetString(key string) string {
	return Config.GetString(key)
}

// GetInt 获取整数配置
func GetInt(key string) int {
	return Config.GetInt(key)
}

// GetBool 获取布尔配置
func GetBool(key string) bool {
	return Config.GetBool(key)
}
