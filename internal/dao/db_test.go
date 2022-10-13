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


func TestDao_SelectMaxAndMin(t *testing.T) {

	mock := newMockDatabase()
	mock.ExpectQuery("SELECT max(id) as max ,min(id) as min FROM t").
		WillReturnRows(sqlmock.NewRows([]string{"min", "max"}).AddRow(1, 10000))

	c := d.SelectMaxAndMin("SELECT max(id) as max ,min(id) as min FROM t")

	assert.Equal(t, c.Max, 10000)
	assert.Equal(t, c.Min, 1)
}

func TestDao_GetClient(t *testing.T) {
	c := d.GetClient()
	assert.NotNil(t, c)
}

func TestDao_ResultTostring(t *testing.T) {


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

func TestJSONSerializer_Scan_Value(t *testing.T) {
	mock := newMockDatabase()

	type Payloads struct {
		Id   int    `json:"id"`
		Type string `json:"type"`
		EsIndexName string `json:"name"`
	}

	type Jobs struct {
		ID uint `gorm:"primaryKey"`
		Payload   Payloads `gorm:"serializer:json"`
	}

	mock.ExpectQuery("SELECT * FROM `jobs` ORDER BY `jobs`.`id` LIMIT 1").
		WillReturnRows(sqlmock.NewRows(
			[]string{"id", "payload"}).
			AddRow(2, `{"id": 41768,"type":"update","name":"t1"}`))

	job := Jobs{}
	d.GetClient().First(&job)
	assert.Equal(t, int(job.ID), 2)
	assert.Equal(t, job.Payload.EsIndexName, "t1")


	mock.ExpectBegin()
	//todo 这里有错空了改
	mock.ExpectExec("UPDATE `jobs` SET `payload`=? WHERE `id` = ?").
		WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()
	job.Payload.Id = 111
	result := d.GetClient().Save(&job)

	assert.NoError(t, result.Error)
	assert.Equal(t, int(result.RowsAffected),1)

}

func newMockDatabase() (sqlmock.Sqlmock) {

	sqlDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)

	if err != nil {
		log.Fatalf("[sqlmock new] %s", err)
	}


	dialector := mysql.Open("127.0.0.1:3306")
	err = d.NewDao(dialector)

	if err == nil {
		panic(any("error new dialector"))
	}

	dialector = mysql.New(mysql.Config{
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