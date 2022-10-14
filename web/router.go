package web

import (
	"github.com/gin-gonic/gin"
	"github.com/zhangjunjie6b/mysql2elasticsearch/web/controllers"
)

func Start() *gin.Engine{
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Static("/bootstrap-4.6.0-dist", "../../web/public/bootstrap-4.6.0-dist/")
	r.LoadHTMLGlob("../../web/views/*")
	r.GET("/", controllers.Index)
	r.POST("/push", controllers.Push)
	r.GET("/progress", controllers.Progress)
	return r
}
