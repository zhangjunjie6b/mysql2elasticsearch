package push

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"github.com/olivere/elastic/v7"
	"main/config"
	"main/service"
	"main/service/monitor"
	"main/service/parse"
	"strconv"
	"strings"
	"sync"
	"time"
)

type section struct {
	Min int
	Max int
}


func BulkPushRun(esConfig service.EsConfig, name string,
	conn config.Content, channel int, configName string) error {

	// 最小 最大 启动数 计算区间
	var max,min int

	db,error := service.NewMysqlObj(conn.Reader.Parameter.Connection.JdbcUrl)

	if error != nil {
		return error
	}

	defer db.Close()

	db.QueryRow(conn.Reader.Parameter.Connection.BoundarySql).Scan(&min,&max)

	channelData := generate(max, min, channel)


	channelWorkNumbers := (max-min)/conn.Writer.Parameter.BatchSize
	monitor.ProgressBars[configName] = monitor.ProgressBar{Total: channelWorkNumbers, Progress: &sync.Map{}}
	monitor.ProgressBars[configName].Progress.Store("number", -1)

	//fmt.Printf("max:%s, min:%s, = : %s \n", max, min, channelWorkNumbers)



	var wg = sync.WaitGroup{}


/*	go func() {
		for true {
			a,_ := monitor.ProgressBars[name].Progress.Load("number")
			fmt.Printf("【%s】total:%s, ing:%s \n",name,monitor.ProgressBars[name].Total, a)
			time.Sleep(time.Second*10)
		}

	}()*/

	for _,v := range channelData{
		wg.Add(1)
		//fmt.Printf("最小:%s, 最大:%s \n",v.Min, v.Max)
		go workProcess(v.Max, v.Min, conn, esConfig, name, &wg, configName)
	}

	wg.Wait()

	return nil
}


/**

 */
func generate(max int, min int, channel int) map[int]section {

	channelPip := make(map[int]section)
	extent :=  (max - min) / channel

	for i:=1; i <= channel; i++ {
		channelPip[i] = section{min,min + extent}
		min = min + extent
	}
	return channelPip
}

/**
	协成分发任务
 */
func workProcess (max int , min int, conn config.Content,
	esConfig service.EsConfig, name string,
	wg *sync.WaitGroup,
	configName string,
	) (error) {

	client := service.NewEsObj(esConfig)

	p, error := client.BulkProcessor().Name("MyBackgroundWorker-1").
		Workers(2).
		BulkActions(conn.Writer.Parameter.BatchSize).
		BulkSize(2 << 20). //2MB
		FlushInterval(30*time.Second).
		Do(context.Background())

	if error != nil {
		fmt.Println(error)
		return error
	}

	db,error := service.NewMysqlObj(conn.Reader.Parameter.Connection.JdbcUrl)

	if error != nil {
		fmt.Println(error)
		return error
	}

	//stmtOut,error := db.Prepare(conn.Reader.Parameter.Connection.QuerySql)

	/*if error != nil {
		fmt.Println(error)
		return error
	}*/

	for i:=min; i <= max; i = i + conn.Writer.Parameter.BatchSize {

		a,_ := monitor.ProgressBars[configName].Progress.Load("number")
		monitor.ProgressBars[configName].Progress.Store("number", a.(int)+1)

		temp := conn.Writer.Parameter.BatchSize + i


		if temp > max {
			temp = max
		}

		//获取数据值和类型
		start := time.Now()
		//some func or operation

		sqlQuery := strings.Replace(conn.Reader.Parameter.Connection.QuerySql, "?", strconv.Itoa(i), 1)
		sqlQuery = strings.Replace(sqlQuery, "?", strconv.Itoa(temp), 1)
		rows, error := db.Query(sqlQuery)

		cost := time.Since(start)
		fmt.Printf("[%d -%d] cost=[%s] \n",i,temp,cost)

		if error != nil {
			fmt.Println(error)
			return error
		}

		columns, error := rows.Columns()
		if error != nil {
			fmt.Println(error)
			return error
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
				return  error
			}

			var value string
			var isID = struct {
				status bool
				key string
				value string
			}{}

			jsonObj := gabs.New()

			for i, col := range values {
				if col != nil {
					value = string(col)
				}

				columnType,error := parse.TypeMapping(columns[i], conn.Writer.Parameter.Column)
				if error != nil {
					fmt.Println(error)
					return  error
				}

				values,error := parse.StrConversion(columnType.Mold, value)
				if error != nil {
					fmt.Println(error)
					return  error
				}

				jsonObj.Set(values, columns[i])

				if columnType.IsID {
					isID.status = true
					isID.key, isID.value =  columns[i], strconv.Itoa(values.(int))
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
		wg.Done()
		p.Close()
		db.Close()
	}()

	return  nil
}