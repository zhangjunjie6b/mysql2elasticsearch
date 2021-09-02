package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Title string
	Port int
	JobList []JobList
}

type JobList struct {
	Name string
	FilePath string
}

func NewConfig () Config{

	config := viper.New()
	config.SetConfigName("config.json")
	config.SetConfigType("json")
	config.AddConfigPath("./config")
	err := config.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	configjson := Config{}
	config.Unmarshal(&configjson)
	return configjson
}

