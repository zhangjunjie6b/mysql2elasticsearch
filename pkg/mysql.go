package service

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)


func  NewMysqlObj(dataSourceName string) (*sql.DB,error) {

	 client,err := sql.Open("mysql", dataSourceName)
	 return  client,err
}


