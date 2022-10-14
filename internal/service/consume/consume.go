package consume

import (
	"fmt"
	"gorm.io/gorm"
	"main/configs"
	"main/internal/mode"
	"time"
)
type ConsumeQueue struct {
	dao *gorm.DB
}
func (d *ConsumeQueue) SetDao(dao *gorm.DB){
	d.dao = dao
}

func (d *ConsumeQueue) Do(synchronousConfig configs.SynchronousConfig)  {

	//todo 临时判断现在队列只有做同步，后面修改配置关系
	if synchronousConfig.Job.Content.Reader.Parameter.Connection.Increment == "" {
		fmt.Printf("%s 未配置Increment \n", synchronousConfig.Job.Content.Writer.Parameter.Index)
		return
	}

	increment := Increment{}
	increment.Init()

	go func() {
		fmt.Printf("%s 开始监听Increment \n",synchronousConfig.Job.Content.Writer.Parameter.Index)
		for true {
			d.run("increment", increment)
			time.Sleep(5*time.Second)
		}
	}()

}

func (d *ConsumeQueue) run (queueName string , consume ConsumeInterface) {
	var jobs []mode.Jobs
	d.dao.Table("push_jobs").
		Where("queue = ? AND del = '0' AND attempts <= 6", queueName).
		FindInBatches(&jobs, 100, func(tx *gorm.DB, batch int) error {

			for k,v := range jobs {

				err := consume.Handle(v)

				if err != nil {
					jobs[k].Attempts = v.Attempts + 1
					jobs[k].LastError = err.Error()
				} else {
					jobs[k].Del = "1"
				}

			}

			tx.Table("push_jobs").Save(&jobs)
			return nil
		})

}