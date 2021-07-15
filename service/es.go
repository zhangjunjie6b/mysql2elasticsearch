package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/olivere/elastic/v7"
	"main/service/errno"
	"strings"
)

type EsConfig struct {
	Addresses string
	Username string
	Password string
}

type SettingsIndexInfo struct {
	Creation_date string
	Number_of_replicas string
	Number_of_shards string
	Provided_name string
	AliaseName map[string]interface{}
	Uuid string
	Version  map[string]string
}

type IndexStatus struct {

	IndexName string
	AliaseName string
	PlanIndexA bool
	PlanIndexB bool

}


// map[creation_date:1610278432105 number_of_replicas:1 number_of_shards:1 provided_name:yandex uuid:J73-HO52Tjik7kbbEBvYDA version:map[created:7070199]]
func NewEsObj(config EsConfig) *elastic.Client {

	client,err := elastic.NewClient(
		elastic.SetURL(config.Addresses, config.Addresses) ,
		elastic.SetBasicAuth(config.Username, config.Password),
		elastic.SetSniff(false),
	)

	if (err != nil) {
		panic(fmt.Errorf("NewEsObj Error : %s", err))
	}
	return  client

}


func GetIndexInfo(client *elastic.Client, ctx context.Context, name string) SettingsIndexInfo {

	info,err := client.IndexGet(name).Do(ctx)

	if err != nil {
		err = fmt.Errorf("[Index-%s]：[%s]", name, err)
		panic(err)
	}

	if len(info) > 1{
		panic(fmt.Errorf("[Index-%s]:[%s]", name, errno.SysAliasExceedLimit))
	}

	for k,_ :=  range info {
		name = k
	}

	c := info[name].Settings["index"]

	index, _ := json.Marshal(&c)
	m := SettingsIndexInfo{}
	err = json.Unmarshal(index, &m)
	m.AliaseName = info[name].Aliases

	if (err != nil) {
		panic(fmt.Errorf("[Index-%s]:[%s]" , name, errno.SysIndexGetInfoTransitionJsonError))
	}

	return m
}




/**
	type IndexStatus struct {
		IndexName string
		AliaseName string
		PlanIndexA bool
		PlanIndexB bool
	}

		IndexName 为空  				表示尚未场景索引，走首次推送流程
		IndexName 有值 AliaseName 为空	表示有索引，但是非程序生成，报错
		IndexName 有值 AliaseName 有值   PlanIndexA&PlanIndexB 其中一个为true    表示非首次运行
		PlanIndexA&PlanIndexB  为true   表示程序正在运行

 */
func GetIndexStatus(client *elastic.Client, ctx context.Context, name string) (IndexStatus, error) {

	var state =  IndexStatus{IndexName: name, AliaseName: "", PlanIndexA: false, PlanIndexB: false}

	exist,_ := client.IndexExists(name).Do(ctx)
	//20 65   45
	if !exist {
		state.IndexName = ""
		return state,nil
	}

	info := GetIndexInfo(client, ctx, name)

	if len(info.AliaseName) > 1 {
		//todo 别名暂时只支持自身维护，多别名场景尚未涉及
		return  state, errors.New(errno.SysIndexAliasExceedLimit)
	}


	state.IndexName = info.Provided_name

	if len(info.AliaseName) == 0 {
		return state,nil
	}

	for k,_ :=  range info.AliaseName{
		state.AliaseName = k
	}


	suffix := strings.Split(state.IndexName, "_")

	if len(suffix) <= 1 {
		return  state, errors.New(errno.SysIndexNameStandardLimit)
	}

	suffix_value := suffix[len(suffix)-1]


	switch suffix_value {
	case "a":
		state.PlanIndexA = true
	case "b":
		state.PlanIndexB = true
	}

	return state, nil
}