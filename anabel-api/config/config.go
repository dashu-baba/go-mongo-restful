package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var config *viper.Viper

// Init initiates configuration
func Init() {
	var err error
	config = viper.New()
	config.SetConfigType("yaml")
	config.SetConfigName("config")
	config.AddConfigPath("config/")
	err = config.ReadInConfig()
	if err != nil {
		log.Fatal("error on parsing configuration file", err)
	}
}

// GetConfig return viper
func GetConfig() *viper.Viper {
	return config
}
