package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"main/configs"
	"main/internal/dao"
	"main/internal/service/consume"
	"main/web"
	"strconv"
)

func main() {

	r := web.Start()
	config := configs.NewConfig()


	for _,v := range config.JobList {
		qu := consume.ConsumeQueue{}
		synchronous,ok := configs.JobNameGetSynchronousConfig(v.Name)

		if ok {
			dao := dao.Dao{}
			err := dao.NewDao(mysql.Open(synchronous.Job.Content.Reader.Parameter.Connection.JdbcUrl))
			if err != nil {
				panic(any(err))
			}
			qu.SetDao(dao.GetClient())
			qu.Do(synchronous)

		} else {
			fmt.Printf("%s JobNameGetSynchronousConfig error \n", v.Name)
		}
	}



	fmt.Printf("web GUI start: http://ip:%d \n", config.Port)

	err := r.Run(":" + strconv.Itoa(config.Port))

	if (err != nil) {
		panic(any(fmt.Errorf("http run error:%s \n", err)))
	}

	
}
