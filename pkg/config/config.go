package config

import (
	"github.com/spf13/viper"
	"log"
)

var config *viper.Viper

func Init() {
	config = viper.New()
	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.AddConfigPath("./config")

	err := config.ReadInConfig()
	if err != nil {
		log.Fatal("Error reading config file: ", err)
	}
}

func GetString(key string) string {
	return config.GetString(key)
}

func GetInt(key string) int {
	return config.GetInt(key)
}

func GetBool(key string) bool {
	return config.GetBool(key)
}
