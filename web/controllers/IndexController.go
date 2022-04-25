package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"main/configs"
	"main/pkg"
	"main/pkg/errno"
	"main/pkg/monitor"
	"main/pkg/push"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func Index(c *gin.Context) {

	configFile := configs.NewConfig()

	job := []map[string]string{}

	ctx := context.Background()
	defer ctx.Done()

	for _, v := range configFile.JobList {
		NewSynchronousConfig := configs.NewSynchronousConfig(v.FilePath)
		esCofig := configs.GetEsConfig(NewSynchronousConfig.Job)
		client := pkg.NewEsObj(esCofig)

		exists, _ := client.IndexExists(NewSynchronousConfig.Job.Content.Writer.Parameter.Index).Do(ctx)

		indexInfo := pkg.SettingsIndexInfo{Uuid: "未创建", Number_of_replicas: "0", Number_of_shards: "0"}

		if exists {
			indexInfo = pkg.GetIndexInfo(ctx, client, NewSynchronousConfig.Job.Content.Writer.Parameter.Index)
		}

		job = append(job, map[string]string{
			"Name":               v.Name,
			"Config_index_name":  NewSynchronousConfig.Job.Content.Writer.Parameter.Index,
			"Index_name":         indexInfo.Provided_name,
			"Number_of_shards":   indexInfo.Number_of_shards,
			"Number_of_replicas": indexInfo.Number_of_replicas,
			"Uuid":               indexInfo.Uuid,
		})

	}

	c.HTML(200, "index.tmpl", gin.H{
		"title":   configFile.Title,
		"jobList": job,
	})
}

func Push(c *gin.Context) {

	name, err := c.GetPostForm("name")

	//没post参数
	if err == false {
		c.JSON(200, gin.H{
			"code":    errno.PushNoName.Code,
			"message": errno.PushNoName.Message,
		})
		return
	}

	esConfig, synchronousConfig, err := configs.JobNameGetESConfig(name)

	//没对应配置文件
	if err == false {
		c.JSON(200, gin.H{
			"code":    errno.PushNoFile.Code,
			"message": errno.PushNoFile.Message,
		})
		return
	}

	client := pkg.NewEsObj(esConfig)
	ctx := context.Background()
	defer ctx.Done()
	//Index当前状态
	state, state_err := pkg.GetIndexStatus(ctx,
		client,
		synchronousConfig.Job.Content.Writer.Parameter.Index)

	if state_err != nil {
		c.JSON(200, gin.H{
			"code":    errno.PushGetIndexStatus.Code,
			"message": state_err.Error(),
		})
		panic(fmt.Errorf("[indexController-Push]:[%s]", state_err))
		return
	}

	//有Index无别名，不让push
	if state.AliaseName == "" && state.IndexName != "" {
		c.JSON(200, gin.H{
			"code":    errno.PushIndexExist.Code,
			"message": errno.PushIndexExist.Message,
		})

		return
	}

	//未创建，首次push
	if state.IndexName == "" {
		//1. 创建index_a

		_, create_err := client.CreateIndex(synchronousConfig.Job.Content.Writer.Parameter.Index + "_a").
			Body(synchronousConfig.Job.Content.Writer.Parameter.Dsl).Do(ctx)

		if create_err != nil {
			c.JSON(200, gin.H{
				"code":    errno.PushCreateIndex.Code,
				"message": create_err.Error(),
			})
			return
		}

		//2. 推送数据
		db_error := push.BulkPushRun(esConfig,
			synchronousConfig.Job.Content.Writer.Parameter.Index+"_a",
			synchronousConfig.Job.Content,
			synchronousConfig.Job.Setting.Speed.Channel,
			name,
		)

		if db_error != nil {
			c.JSON(200, gin.H{
				"code":    errno.PushError.Code,
				"message": db_error.Error(),
			})
			client.DeleteIndex(synchronousConfig.Job.Content.Writer.Parameter.Index + "_a").Do(ctx)
			return
		}
		//3. 创建别名

		client.Alias().Add(synchronousConfig.Job.Content.Writer.Parameter.Index+"_a",
			synchronousConfig.Job.Content.Writer.Parameter.Index).Do(ctx)

		//非首次push
	} else if state.PlanIndexA || state.PlanIndexB {

		var now_index_suffix string
		var new_index_suffix string

		//1. 获取当前状态
		if state.PlanIndexA {
			new_index_suffix = synchronousConfig.Job.Content.Writer.Parameter.Index + "_b"
			now_index_suffix = synchronousConfig.Job.Content.Writer.Parameter.Index + "_a"
		} else if state.PlanIndexB {
			new_index_suffix = synchronousConfig.Job.Content.Writer.Parameter.Index + "_a"
			now_index_suffix = synchronousConfig.Job.Content.Writer.Parameter.Index + "_b"
		}

		//2. 创建索引 a || b
		_, create_err := client.CreateIndex(new_index_suffix).
			Body(synchronousConfig.Job.Content.Writer.Parameter.Dsl).Do(ctx)

		if create_err != nil {
			c.JSON(200, gin.H{
				"code":    errno.PushCreateIndex.Code,
				"message": create_err.Error(),
			})
			return
		}

		//3. 推送数据
		db_error := push.BulkPushRun(esConfig,
			new_index_suffix,
			synchronousConfig.Job.Content,
			synchronousConfig.Job.Setting.Speed.Channel,
			name,
		)

		if db_error != nil {
			c.JSON(200, gin.H{
				"code":    errno.PushError.Code,
				"message": db_error.Error(),
			})
			client.DeleteIndex(new_index_suffix).Do(ctx)
			return
		}

		//4. 别名调换

		client.Alias().Add(new_index_suffix,
			synchronousConfig.Job.Content.Writer.Parameter.Index).Do(ctx)

		//5. 删除老index
		delete(monitor.ProgressBars, name)

		client.DeleteIndex(now_index_suffix).Do(ctx)

		c.JSON(200, gin.H{
			"code":    200,
			"message": "ok",
		})
	}

}

func Progress(c *gin.Context) {

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil) // 升级协议

	defer conn.Close()

	if err != nil {
		log.Println(err)
		return
	}

	for {
		messageType, message, err := conn.ReadMessage()

		if err != nil {
			log.Println(err)
			return
		}

		var progressBarJson []byte
		var IsClose = false
		for {
			progress, ok := monitor.ProgressBars[string(message[:])]

			if ok == true {
				progress := progress.Progress

				progressBarJson, _ = json.Marshal(monitor.ProgressBarJson{
					Total:    monitor.ProgressBars[string(message[:])].Total,
					Progress: progress,
					Name:     string(message[:]),
					Status:   101,
				})

			} else {

				progressBarJson, _ = json.Marshal(monitor.ProgressBarJson{
					Name:   string(message[:]),
					Status: 200,
				})

				IsClose = true
			}

			if err := conn.WriteMessage(messageType, progressBarJson); err != nil {
				log.Println(err)
				return
			}

			if IsClose {
				time.Sleep(time.Second * 60)
				conn.Close()
			}

			time.Sleep(time.Second * 2)
		}

	}

}

func Restart(c *gin.Context) {

	//todo 重新启动还要找其他的解决方法
	str, _ := os.Getwd()
	fmt.Println(str)
	s := exec.Command("bash", "-c", "kill `lsof -t -i:9100` && "+str+"/main")
	output, _ := s.CombinedOutput()
	fmt.Println(string(output))

}
