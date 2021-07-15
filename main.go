package main

import (
	"main/webGUI"
)

func main()  {
	/*StaticConfig := config.NewConfig()*/
	/*service.NewEsObj(StaticConfig)
	service.NewMysqlObj(StaticConfig, "0")*/
	webGUI.Start()
}

