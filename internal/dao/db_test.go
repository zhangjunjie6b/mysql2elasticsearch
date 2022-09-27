package dao

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"log"
	"main/configs"
	"main/internal/pkg/errno"
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

func TestResultTostring(t *testing.T) {


	mock := newMockDatabase()
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "keyword","picheight","cl_num"}).
			AddRow(1, "测试", 100, 100).
			AddRow(2, "测试2", 200, 200))


	parameter := []configs.Column{
		{Name: "id" , Type: "id"},
		{Name: "keyword" , Type: "text"},
		{Name: "picheight" , Type: "integer"},
		{Name: "cl_num" , Type: "integer"},
	}


	rows,err := d.client.Raw("SELECT").Rows()
	assert.NoError(t, err)
	re,err := d.ResultTostring(rows, parameter)
	assert.NoError(t, err)

	expect := []ResultJson{
		{`{"cl_num":100,"id":1,"keyword":"测试","picheight":100}`,
			FieldID{true,"id", "1"},
		},
		{`{"cl_num":200,"id":2,"keyword":"测试2","picheight":200}`,
			FieldID{true,"id", "2"},
		},
	}

	assert.Equal(t, re, expect)


	//使用未定义的映射类型
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "keyword","picheight","cl_num"}).
			AddRow(1, "测试", 100, 100).
			AddRow(2, "测试2", 200, 200))

	rows,err = d.client.Raw("SELECT").Rows()
	assert.NoError(t, err)
	parameter = []configs.Column{
		{Name: "id" , Type: "ids"},
		{Name: "keyword" , Type: "text"},
		{Name: "picheight" , Type: "integer"},
		{Name: "cl_num" , Type: "integer"},
	}

	_,err = d.ResultTostring(rows, parameter)

	assert.Equal(t,  fmt.Errorf("[%s]:%s", "id", errno.SysTypeUndefined), err)


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