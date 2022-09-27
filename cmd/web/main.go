package main

import (
	"fmt"
	"main/configs"
	"main/internal/service/consume"
	"main/web"
	"strconv"
)

func main() {

	r := web.Start()
	config := configs.NewConfig()

	for _,v := range config.JobList {
		qu := consume.ConsumeQueue{}
		qu.Do(v.Name)
	}
	
	fmt.Printf("web GUI start: http://ip:%d \n", config.Port)

	//err := endless.ListenAndServe(":" + strconv.Itoa(config.Port), r)
	err := r.Run(":" + strconv.Itoa(config.Port))

	if (err != nil) {
		panic(any(fmt.Errorf("http run error:%s \n", err)))
	}

	
}
