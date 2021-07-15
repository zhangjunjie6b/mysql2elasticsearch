package webGUI

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"main/webGUI/controllers"
)

func Start() {

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Static("/bootstrap-4.6.0-dist", "./webGUI/public/bootstrap-4.6.0-dist/")
	r.LoadHTMLGlob("./webGUI/views/*")
	r.GET("/", controllers.Index)
	r.POST("/push", controllers.Push)

	fmt.Println("web GUI start: http://ip:9100")
	err := r.Run(":9100")

	if (err != nil) {
		panic(fmt.Errorf("http run error:%s \n", err))
	}

}
