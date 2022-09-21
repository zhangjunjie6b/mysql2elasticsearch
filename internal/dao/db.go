package dao

import (
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type Dao struct {
	client *gorm.DB
}

func (d *Dao) GetClient() (*gorm.DB) {
	return d.client
}

func (d *Dao) NewDao(dialector gorm.Dialector) (error){


	 client, err := gorm.Open(dialector, &gorm.Config{})

	 if err!= nil {
	 	return err
	 }
	 d.client = client
	 return  nil
}


type Section struct {
	Max int
	Min int
}

func (d *Dao) SelectMaxAndMin(sql string) Section {
	var section Section
	d.client.Raw(sql).Scan(&section)
	return section
}

