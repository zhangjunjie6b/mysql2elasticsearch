package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"main/configs"
	"main/web/controllers"
	"strconv"
)

func Start() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Static("/bootstrap-4.6.0-dist", "../../web/public/bootstrap-4.6.0-dist/")
	r.LoadHTMLGlob("../../web/views/*")
	r.GET("/", controllers.Index)
	r.POST("/push", controllers.Push)
	r.GET("/progress", controllers.Progress)
	r.GET("/reset", controllers.Restart)


	config := configs.NewConfig()
	fmt.Printf("web GUI start: http://ip:%d \n", config.Port)

	//err := endless.ListenAndServe(":" + strconv.Itoa(config.Port), r)
	err := r.Run(":" + strconv.Itoa(config.Port))

	if (err != nil) {
		panic(fmt.Errorf("http run error:%s \n", err))
	}

}
