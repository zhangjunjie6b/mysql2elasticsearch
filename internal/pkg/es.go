package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/pkg/errno"
	"strings"
)

type EsConfig struct {
	Addresses string
	Username  string
	Password  string
}

type SettingsIndexInfo struct {
	Creation_date      string
	Number_of_replicas string
	Number_of_shards   string
	Provided_name      string
	AliaseName         map[string]interface{}
	Uuid               string
	Version            map[string]string
}

type IndexStatus struct {
	IndexName  string
	AliaseName string
	PlanIndexA bool
	PlanIndexB bool
}

type EsInterface interface {
	NewEsObj(config EsConfig) (*elastic.Client, error)
	GetIndexInfo(name string) (SettingsIndexInfo, error)
	GetIndexStatus(name string) (IndexStatus, error)
}

type ES struct {
	Client *elastic.Client
	Ctx    context.Context
}

// map[creation_date:1610278432105 number_of_replicas:1 number_of_shards:1 provided_name:yandex uuid:J73-HO52Tjik7kbbEBvYDA version:map[created:7070199]]
func (e *ES) NewEsObj(config EsConfig) (*elastic.Client, error) {

	client, err := elastic.NewClient(
		elastic.SetURL(config.Addresses, config.Addresses),
		elastic.SetBasicAuth(config.Username, config.Password),
		elastic.SetSniff(false),
	)

	if err != nil {
		return nil, errors.Wrap(err, "NewEsObj Error")
	}

	e.Client = client
	e.Ctx = context.Background()
	return e.Client, nil

}

func (e *ES) GetIndexInfo(name string) (SettingsIndexInfo, error) {

	info, err := e.Client.IndexGet(name).Do(e.Ctx)

	if err != nil {
		return SettingsIndexInfo{}, fmt.Errorf("[Index-%s]：[%s]", name, err)
	}

	if len(info) > 1 {
		return SettingsIndexInfo{}, fmt.Errorf("[Index-%s]:[%s]", name, errno.SysAliasExceedLimit)
	}

	for k := range info {
		name = k
	}

	c := info[name].Settings["index"]

	index, _ := json.Marshal(&c)
	m := SettingsIndexInfo{}
	err = json.Unmarshal(index, &m)

	if err != nil {
		return SettingsIndexInfo{}, fmt.Errorf("[Index-%s]:[%s]", name, errno.SysIndexGetInfoTransitionJsonError)
	}
	m.AliaseName = info[name].Aliases

	return m, nil
}

/**
type IndexStatus struct {
	IndexName string
	AliaseName string
	PlanIndexA bool
	PlanIndexB bool
}

	IndexName 为空  				表示尚未场景索引，走首次推送流程
	IndexName 有值 AliaseName 为空	表示有索引，但是非程序生成，返回空信息
	IndexName 有值 AliaseName 有值   PlanIndexA&PlanIndexB 其中一个为true    表示非首次运行
	PlanIndexA&PlanIndexB  为true   表示程序正在运行

*/
func (e *ES) GetIndexStatus(name string) (IndexStatus, error) {

	var state = IndexStatus{IndexName: name, AliaseName: "", PlanIndexA: false, PlanIndexB: false}

	exist, err := e.Client.IndexExists(name).Do(e.Ctx)

	if !exist {
		state.IndexName = ""
		return state, nil
	}

	info, err := e.GetIndexInfo(name)

	if err != nil {
		//todo 多别名场景尚未涉及
		return state, err
	}

	state.IndexName = info.Provided_name

	if len(info.AliaseName) == 0 {
		return state, nil
	}

	for k := range info.AliaseName {
		state.AliaseName = k
	}

	suffix := strings.Split(state.IndexName, "_")
	suffix_value := suffix[len(suffix)-1]

	switch suffix_value {
	case "a":
		state.PlanIndexA = true
	case "b":
		state.PlanIndexB = true
	}

	return state, nil
}
