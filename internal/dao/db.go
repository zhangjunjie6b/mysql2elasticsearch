package dao

import (
	"database/sql"
	"github.com/Jeffail/gabs/v2"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"main/configs"
	"main/internal/pkg/parse"
	"os"
	"strconv"
	"time"
)

type Dao struct {
	client *gorm.DB
}

type ResultJson struct {
	JsonString string
	FieldID FieldID
}

type FieldID struct {
	Status bool
	Key    string
	Value  string
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
	 schema.RegisterSerializer("json", JSONSerializer{})

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


func (d *Dao) ResultTostring(rows *sql.Rows, configColumn []configs.Column) ([]ResultJson,error) {

	resultJson := []ResultJson{}
	columns, err := rows.Columns()
	if err != nil {
		return resultJson,err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return resultJson, err
		}

		var value string
		var isID FieldID

		jsonObj := gabs.New()

		for i, col := range values {
			if col != nil {
				value = string(col)
			}

			columnType, err := parse.TypeMapping(columns[i], configColumn)
			if err != nil {
				return resultJson, err
			}

			values, err := parse.StrConversion(columnType.Mold, value)
			if err != nil {
				return resultJson, err
			}

			jsonObj.Set(values, columns[i])

			if columnType.IsID {
				isID.Status = true
				isID.Key, isID.Value = columns[i], strconv.Itoa(values.(int))
			}
		}

		resultJson = append(resultJson,ResultJson{
			JsonString: jsonObj.String(),
			FieldID:   isID,
		})

	}

	return resultJson, nil
}