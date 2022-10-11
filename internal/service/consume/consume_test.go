package consume

import (
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"log"
	"main/configs"
	"main/internal/dao"
	"main/internal/mode"
	"testing"
)

var d dao.Dao

func TestDo(t *testing.T) {
	queue := ConsumeQueue{}
	config := configs.SynchronousConfig{}
	queue.Do(config)
	config = configs.SynchronousConfig{
			Job: struct {
				Setting configs.Setting
				Content configs.Content
			}{Content: configs.Content{Reader: struct {
				Name      string
				Parameter configs.ReaderParameter
			}{Parameter: configs.ReaderParameter{
				Connection: configs.Connection{Increment: "select"},
			}}}},
	}
	newMockDatabase()
	queue.SetDao(d.GetClient())
	queue.Do(config)
}

type TConsume struct {
	t *testing.T
}


func (c TConsume) Handle(data interface{}) error{
	da,_ := data.(mode.Jobs)
	if da.ID == 1 {
		return errors.New("err")
	}
	return nil
}

func TestRun(t *testing.T) {
	queue := ConsumeQueue{}
	mock := newMockDatabase()
	sql := "SELECT * FROM `push_jobs` WHERE queue = ? AND del = '0' AND attempts <= 6 ORDER BY `push_jobs`.`id` LIMIT 100"

	mock.ExpectQuery(sql).WillReturnRows(
				sqlmock.NewRows([]string{
					"id","queue","payload","del","attempts","lastError",
				}).
					AddRow(1,"increment",`{"id": 1,"type":"update","name":"t1"}`,"0",1,"").
					AddRow(2,"increment",`{"id": 2,"type":"update","name":"t2"}`,"0",0,""))

	mock.ExpectBegin()
	sql = "INSERT INTO `push_jobs` (`queue`,`payload`,`del`,`attempts`,`last_error`,`id`) VALUES (?,?,?,?,?,?),(?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `queue`=VALUES(`queue`),`payload`=VALUES(`payload`),`del`=VALUES(`del`),`attempts`=VALUES(`attempts`),`last_error`=VALUES(`last_error`)"

	row1 := mode.Payloads{
		Id:          1,
		Type:        "update",
		EsIndexName: "t1",
	}

	row2 := mode.Payloads{
		Id:          2,
		Type:        "update",
		EsIndexName: "t2",
	}

	j1,_ := json.Marshal(row1)
	j2,_ := json.Marshal(row2)

	mock.ExpectExec(sql).
		WithArgs("increment",j1,"0",2,"err",1,
						"increment",j2,"1",0,"",2).
		WillReturnResult(sqlmock.NewResult(2, 2))
	mock.ExpectCommit()

	tConsume := TConsume{}
	tConsume.t = t
	queue.SetDao(d.GetClient())
	queue.run("increment", tConsume)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

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