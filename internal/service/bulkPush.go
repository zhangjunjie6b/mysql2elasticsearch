package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/olivere/elastic/v7"
	"github.com/zhangjunjie6b/mysql2elasticsearch/configs"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/dao"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/pkg"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/pkg/monitor"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/pkg/parse"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Bulk struct {
	bulkPushRunWg sync.WaitGroup
	dao           dao.Dao
	es            pkg.ES
	config        configs.Content
}

func (b *Bulk) Init(conn configs.Content, d dao.Dao, es pkg.ES) error {
	b.config = conn
	b.dao = d
	b.es = es
	b.bulkPushRunWg = sync.WaitGroup{}
	return nil
}

func (b *Bulk) Run(channelData map[int]dao.Section, indexname string) error {

	channelWorkNumbers := (channelData[len(channelData)].Max - channelData[1].Min) / b.config.Writer.Parameter.BatchSize
	monitor.ProgressBars[b.config.Writer.Parameter.Index] = &monitor.ProgressBar{Total: channelWorkNumbers, Progress: 0}

	for _, v := range channelData {
		b.bulkPushRunWg.Add(1)
		go b.workProcess(v, indexname)
	}

	b.bulkPushRunWg.Wait()

	return nil
}

func (b *Bulk) workProcess(section dao.Section, indexname string) error {

	client := b.es.Client
	var p *elastic.BulkProcessor
	var error error
	var timesCount = 0
	//设置批量提交
	for true {
		p, error = client.BulkProcessor().Name("MyBackgroundWorker-1").
			Workers(2).
			BulkActions(b.config.Writer.Parameter.BatchSize).
			BulkSize(2 << 20). //2MB
			FlushInterval(30 * time.Second).
			Do(context.Background())

		if error != nil {
			if timesCount > 6 {
				panic(any(error)) // 需要人工介入直接挂服务
			}
			time.Sleep(time.Second * 10)
			timesCount++
		} else {
			break
		}
	}

	for i := section.Min; i <= section.Max; i = i + b.config.Writer.Parameter.BatchSize {

		monitor.ProgressBars[b.config.Writer.Parameter.Index].M.Lock()
		monitor.ProgressBars[b.config.Writer.Parameter.Index].Progress += 1
		monitor.ProgressBars[b.config.Writer.Parameter.Index].M.Unlock()
		temp := b.config.Writer.Parameter.BatchSize + i

		if temp > section.Max {
			temp = section.Max
		}

		//获取数据值和类型
		start := time.Now()
		sqlQuery := strings.Replace(b.config.Reader.Parameter.Connection.QuerySql, "?", strconv.Itoa(i), 1)
		sqlQuery = strings.Replace(sqlQuery, "?", strconv.Itoa(temp), 1)

		var rows *sql.Rows
		timesCount = 0

		for true {

			rows, error = b.dao.GetClient().Raw(sqlQuery).Rows()

			if error != nil {
				if timesCount > 10 {
					panic(any(error))
				}
				time.Sleep(time.Second * 10)
				fmt.Printf("query error : %d", timesCount)
				timesCount++
			} else {
				break
			}
		}

		cost := time.Since(start)
		fmt.Printf("[%d -%d] cost=[%s] \n", i, temp, cost)

		columns, error := rows.Columns()
		if error != nil {
			fmt.Println(error)
			panic(any(error))
		}

		values := make([]sql.RawBytes, len(columns))
		scanArgs := make([]interface{}, len(values))

		for i := range values {
			scanArgs[i] = &values[i]
		}

		for rows.Next() {
			error := rows.Scan(scanArgs...)
			if error != nil {
				fmt.Println(error)
				panic(any(error))
			}

			var value string
			var isID = struct {
				status bool
				key    string
				value  string
			}{}

			jsonObj := gabs.New()

			for i, col := range values {
				if col != nil {
					value = string(col)
				}

				columnType, error := parse.TypeMapping(columns[i], b.config.Writer.Parameter.Column)
				if error != nil {
					fmt.Println(error)
					panic(any(error))
				}

				values, error := parse.StrConversion(columnType.Mold, value)
				if error != nil {
					fmt.Println(error)
					panic(any(error))
				}

				if columnType.Mold == "vector" {

					jsonObjParsed, err := gabs.ParseJSON([]byte(values.(string)))
					if err != nil {

						fmt.Println("Error parsing JSON:", err)
						//为项目组指定修改
						//向量搜索会出现 0,解析失败的情况
						//把string中的0.替换成00000
						values = strings.Replace(values.(string), "0.", "0.00000", -1)
						jsonObjParsed, err = gabs.ParseJSON([]byte(values.(string)))
						if err != nil {
							fmt.Println("Error parsing JSON retry:", err)
						}
					}
					jsonObj.Set(jsonObjParsed.Data(), columns[i])

				} else {
					jsonObj.Set(values, columns[i])
				}

				if columnType.IsID {
					isID.status = true
					isID.key, isID.value = columns[i], strconv.Itoa(values.(int))
				}

			}

			var r *elastic.BulkIndexRequest

			if isID.status {
				r = elastic.NewBulkIndexRequest().Index(indexname).Type("_doc").Id(isID.value).Doc(jsonObj.String())
			} else {
				r = elastic.NewBulkIndexRequest().Index(indexname).Type("_doc").Doc(jsonObj.String())
			}

			p.Add(r)
		}

	}
	defer func() {
		b.bulkPushRunWg.Done()
	}()

	return nil
}

func (b *Bulk) Generate(section dao.Section, chanNumber int) map[int]dao.Section {

	channelPip := make(map[int]dao.Section)
	extent := ((section.Max - section.Min) / chanNumber) + 1

	for i := 1; i <= chanNumber; i++ {
		channelPip[i] = dao.Section{section.Min, section.Min + extent}
		section.Min = section.Min + extent
	}

	return channelPip

}
