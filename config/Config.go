package config

import (
	"fmt"
	"github.com/spf13/viper"
	"main/service"
	"main/service/errno"
	"time"
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
	Loop:
	config := viper.New()
	config.SetConfigName("config.json")
	config.SetConfigType("json")
	config.AddConfigPath("./config")
	err := config.ReadInConfig()

	if err != nil {
		fmt.Println(errno.SysConfigNotFind)
		fmt.Println("正在生成默认配置文件")
		makeDefaultConfig()
		time.Sleep( 15 * time.Second)
		goto Loop
	}

	configjson := Config{}
	config.Unmarshal(&configjson)
	return configjson
}

func makeDefaultConfig()  {
	service.CopyFile("./config/config.json", "./webGUI/default/config/config.json")
	service.CopyFile("./config/example.json", "./webGUI/default/config/example.json")
}
