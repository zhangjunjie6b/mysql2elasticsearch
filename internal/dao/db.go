package dao

import (
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

type Dao struct {
	client *gorm.DB
}

func (d *Dao) GetClient() (*gorm.DB) {
	return d.client
}

func (d *Dao) NewDao(dialector gorm.Dialector) (error){


	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,   // 慢 SQL 阈值
			LogLevel:      logger.Error, // 日志级别
			IgnoreRecordNotFoundError: true,   // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:      false,         // 禁用彩色打印
		},
	)

	 client, err := gorm.Open(dialector, &gorm.Config{Logger: newLogger})

	 if err!= nil {
	 	return err
	 }
	 d.client = client
	 return  nil
}

type Section struct {
	Min int
	Max int
}

func (d *Dao) SelectMaxAndMin(sql string) Section {
	var section Section
	d.client.Raw(sql).Scan(&section)
	return section
}
