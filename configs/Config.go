package configs

import (
	"fmt"
	"github.com/spf13/viper"
	"main/internal/pkg"
	"main/internal/pkg/errno"
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
	var  errNumber int = 0
	Loop:
	config := viper.New()
	config.SetConfigName("config.json")
	config.SetConfigType("json")
	config.AddConfigPath("../../configs")
	err := config.ReadInConfig()

	if err != nil {
		errNumber ++
		fmt.Println(errno.SysConfigNotFind)
		fmt.Println("正在生成默认配置文件")
		makeDefaultConfig()
		time.Sleep( 5 * time.Second)

		if errNumber >3 {
			panic(any("配置文件检查出错"))
		}

		goto Loop
	}
	configjson := Config{}
	config.Unmarshal(&configjson)
	return configjson
}

func makeDefaultConfig()  {
	pkg.CopyFile("../../configs/config.json", "../../web/default/config/config.json")
	pkg.CopyFile("../../configs/example.json", "../../web/default/config/example.json")
}
