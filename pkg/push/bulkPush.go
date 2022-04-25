package push

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/olivere/elastic/v7"
	"main/configs"
	"main/pkg"
	"main/pkg/monitor"
	"main/pkg/parse"
	"strconv"
	"strings"
	"sync"
	"time"
)

type section struct {
	Min int
	Max int
}

var BulkPushRunWg = sync.WaitGroup{}

func BulkPushRun(esConfig pkg.EsConfig, name string,
	conn configs.Content, channel int, configName string) error {
	// 最小 最大 启动数 计算区间
	var max, min int

	db, error := pkg.NewMysqlObj(conn.Reader.Parameter.Connection.JdbcUrl)

	if error != nil {
		return error
	}

	defer db.Close()

	db.QueryRow(conn.Reader.Parameter.Connection.BoundarySql).Scan(&min, &max)

	channelData := generate(max, min, channel)

	channelWorkNumbers := (max - min) / conn.Writer.Parameter.BatchSize
	monitor.ProgressBars[configName] = &monitor.ProgressBar{Total: channelWorkNumbers, Progress: 0}


	for _, v := range channelData {
		BulkPushRunWg.Add(1)
		//fmt.Printf("最小:%s, 最大:%s \n",v.Min, v.Max)
		go workProcess(v.Max, v.Min, conn, esConfig, name, &BulkPushRunWg, configName)
	}

	BulkPushRunWg.Wait()

	return nil
}

/**

 */
func generate(max int, min int, channel int) map[int]section {

	channelPip := make(map[int]section)
	extent := ((max - min) / channel) +1

	for i := 1; i <= channel; i++ {
		channelPip[i] = section{min, min + extent}
		min = min + extent
	}
	return channelPip
}

/**
协成分发任务
*/
func workProcess(max int, min int, conn configs.Content,
	esConfig pkg.EsConfig, name string,
	BulkPushRunWg *sync.WaitGroup,
	configName string,
) error {

	client := pkg.NewEsObj(esConfig)

	var p *elastic.BulkProcessor
	var error error
	var timesCount = 0
	for true {
		p, error = client.BulkProcessor().Name("MyBackgroundWorker-1").
			Workers(2).
			BulkActions(conn.Writer.Parameter.BatchSize).
			BulkSize(2 << 20). //2MB
			FlushInterval(30 * time.Second).
			Do(context.Background())

		if error != nil {
			if timesCount > 6 {
				panic(error) // 需要人工介入直接挂服务
			}
			time.Sleep(time.Second * 10)
			timesCount++
		} else {
			break
		}
	}

	db, error := pkg.NewMysqlObj(conn.Reader.Parameter.Connection.JdbcUrl)

	if error != nil {
		fmt.Println(error)
		return error
	}

	for i := min; i <= max; i = i + conn.Writer.Parameter.BatchSize {


	    monitor.ProgressBars[configName].M.Lock()

		monitor.ProgressBars[configName].Progress += 1

		monitor.ProgressBars[configName].M.Unlock()

		temp := conn.Writer.Parameter.BatchSize + i

		if temp > max {
			temp = max
		}

		//获取数据值和类型
		start := time.Now()
		//some func or operation

		sqlQuery := strings.Replace(conn.Reader.Parameter.Connection.QuerySql, "?", strconv.Itoa(i), 1)
		sqlQuery = strings.Replace(sqlQuery, "?", strconv.Itoa(temp), 1)

		var rows *sql.Rows
		timesCount = 0

		for true {
			rows, error = db.Query(sqlQuery)

			if error != nil {
				if timesCount > 6 {
					panic(error)
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
			panic(error)
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
				panic(error)
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

				columnType, error := parse.TypeMapping(columns[i], conn.Writer.Parameter.Column)
				if error != nil {
					fmt.Println(error)
					panic(error)
				}

				values, error := parse.StrConversion(columnType.Mold, value)
				if error != nil {
					fmt.Println(error)
					panic(error)
				}

				jsonObj.Set(values, columns[i])

				if columnType.IsID {
					isID.status = true
					isID.key, isID.value = columns[i], strconv.Itoa(values.(int))
				}

			}

			var r *elastic.BulkIndexRequest

			if isID.status {
				r = elastic.NewBulkIndexRequest().Index(name).Type("_doc").Id(isID.value).Doc(jsonObj.String())
			} else {
				r = elastic.NewBulkIndexRequest().Index(name).Type("_doc").Doc(jsonObj.String())
			}

			p.Add(r)
		}

	}
	defer func() {
		BulkPushRunWg.Done()
		p.Close()
		db.Close()
	}()

	return nil
}
