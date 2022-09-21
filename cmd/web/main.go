package main

import (
	"fmt"
	"main/configs"
	"main/web"
	"strconv"
)


func main() {
	r := web.Start()
	config := configs.NewConfig()
	fmt.Printf("web GUI start: http://ip:%d \n", config.Port)

	//err := endless.ListenAndServe(":" + strconv.Itoa(config.Port), r)
	err := r.Run(":" + strconv.Itoa(config.Port))

	if (err != nil) {
		panic(any(fmt.Errorf("http run error:%s \n", err)))
	}
}
