package dao

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"log"
	"testing"
)

var d Dao


func TestSelectMaxAndMin(t *testing.T) {

	mock := newMockDatabase()
	mock.ExpectQuery("SELECT max(id) as max ,min(id) as min FROM t").
		WillReturnRows(sqlmock.NewRows([]string{"min", "max"}).AddRow(1, 10000))

	c := d.SelectMaxAndMin("SELECT max(id) as max ,min(id) as min FROM t")

	assert.Equal(t, c.Max, 10000)
	assert.Equal(t, c.Min, 1)
}

func TestGetClient(t *testing.T) {
	c := d.GetClient()
	assert.NotNil(t, c)
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