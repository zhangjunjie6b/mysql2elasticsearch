package dao

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"log"
	"testing"
)

var d Dao


func TestName(t *testing.T) {

	mock := newMockDatabase()
	mock.ExpectQuery("SELECT max(id) as max ,min(id) as min FROM zjj_articles").
		WillReturnRows(sqlmock.NewRows([]string{"max", "min"}).AddRow(1, 10000))

	c := d.SelectMaxAndMin("SELECT max(id) as max ,min(id) as min FROM zjj_articles")
	fmt.Println(c)
}



func newMockDatabase() (sqlmock.Sqlmock) {

	sqlDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)

	if err != nil {
		log.Fatalf("[sqlmock new] %s", err)
	}


	dialector := mysql.New(mysql.Config{
		Conn: sqlDB,
		DriverName: "mysql",
		SkipInitializeWithVersion: true,
	})

	err = d.NewDao(dialector)

	if err != nil {
		log.Fatalf("[gorm open] %s", err)
	}

	return  mock
}