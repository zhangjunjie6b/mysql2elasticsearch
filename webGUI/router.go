package webGUI

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"main/config"
	"main/webGUI/controllers"
	"strconv"
)

func Start() {

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Static("/bootstrap-4.6.0-dist", "./webGUI/public/bootstrap-4.6.0-dist/")
	r.LoadHTMLGlob("./webGUI/views/*")
	r.GET("/", controllers.Index)
	r.POST("/push", controllers.Push)
	r.GET("/progress", controllers.Progress)
	r.GET("/reset", controllers.Restart)


	config := config.NewConfig()
	fmt.Printf("web GUI start: http://ip:%d \n", config.Port)

	//err := endless.ListenAndServe(":" + strconv.Itoa(config.Port), r)
	err := r.Run(":" + strconv.Itoa(config.Port))

	if (err != nil) {
		panic(fmt.Errorf("http run error:%s \n", err))
	}

}
