package pkg

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/pkg/errno"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockHander struct {
	Resp string
	Path string
	Method []string
	HttpCode int
}


func TestGetIndexInfo(t *testing.T) {
	//mock
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	assert := assert.New(t)
	//test NewEsObj
	handler = mockRouteHandler([]mockHander{
		{
			Resp: `{ "names" : "a709c7505efe", "cluster_name" : "docker-cluster", "cluster_uuid" : "8DC8kaEsQ_ClkJtB3OHy9Q", "version" : { "number" : "7.7.1", "build_flavor" : "default", "build_type" : "docker", "build_hash" : "ad56dce891c901a492bb1ee393f12dfff473a423", "build_date" : "2020-05-28T16:30:01.040088Z", "build_snapshot" : false, "lucene_version" : "8.5.1", "minimum_wire_compatibility_version" : "6.8.0", "minimum_index_compatibility_version" : "6.0.0-beta1" }, "tagline" : "You Know, for Search"}`,
			Path: "/",
			Method: []string {"GET", "HEAD"},
			HttpCode: 200,
		},
		{
			Resp: `{"error":{"root_cause":[{"type":"index_not_found_exception","reason":"no such index [t]","resource.type":"index_or_alias","resource.id":"t","index_uuid":"_na_","index":"t"}],"type":"index_not_found_exception","reason":"no such index [t]","resource.type":"index_or_alias","resource.id":"t","index_uuid":"_na_","index":"t"},"status":404}`,
			Path: "/t",
			Method: []string {"GET", "HEAD"},
			HttpCode: 404,

		},
		{
			Resp: `{"weibo":{"aliases":{},"mappings":{"properties":{"tel":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"wb_id":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}}},"settings":{"index":{"number_of_shards":"1","provided_name":"weibo","max_result_window":"10000","creation_date":"1610209867377","number_of_replicas":"1","uuid":"EEB86YAhQ-mXs5xb3oN-Jg","version":{"created":"7070199"}}}}}`,
			Path: "/weibo",
			Method: []string {"GET", "HEAD"},
			HttpCode: 200,
		},

		{
			Resp: `{"tianya":{"aliases":{"a":{}},"mappings":{"properties":{"mail":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"name":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"pass":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"tags":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}}},"settings":{"index":{"number_of_shards":"1","provided_name":"tianya","max_result_window":"10000","creation_date":"1614246655737","number_of_replicas":"1","uuid":"AjVKnCUlRaCXKJkngb-MIQ","version":{"created":"7070199"}}}},"weibo":{"aliases":{"a":{}},"mappings":{"properties":{"tel":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"wb_id":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}}},"settings":{"index":{"number_of_shards":"1","provided_name":"weibo","max_result_window":"10000","creation_date":"1610209867377","number_of_replicas":"1","uuid":"EEB86YAhQ-mXs5xb3oN-Jg","version":{"created":"7070199"}}}}}`,
			Path: "/a",
			Method: []string {"GET", "HEAD"},
			HttpCode: 200,
		},

	})
	es := ES{}
	config := EsConfig{Addresses: ts.URL}
	_,err := es.NewEsObj(config)
	assert.Nil(err)
	//服务不可用
	config = EsConfig{Addresses: "http://127.0.0.1"}
	_,err = es.NewEsObj(config)
	assert.Error(err)

	//test GetIndexInfo
	//正常情况
	info,err := es.GetIndexInfo("weibo")
	assert.Nil(err)
	assert.Equal("1610209867377", info.Creation_date)
	assert.Equal("1", info.Number_of_replicas)
	assert.Equal("1", info.Number_of_shards)
	assert.Equal("1", info.Number_of_shards)
	assert.Equal("weibo", info.Provided_name)
	assert.Empty(info.AliaseName)
	assert.Equal("EEB86YAhQ-mXs5xb3oN-Jg",info.Uuid)
	assert.Equal(map[string]string {"created":"7070199"}, info.Version)
	//索引不存在
	_,err = es.GetIndexInfo("t")
	assert.Error(err)
	//多别名
	_,err = es.GetIndexInfo("a")
	assert.Equal(fmt.Errorf("[Index-%s]:[%s]", "a", errno.SysAliasExceedLimit), err)


}

func TestGetIndexStatus(t *testing.T)  {

	//mock
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	assert := assert.New(t)
	//test NewEsObj
	handler = mockRouteHandler([]mockHander{
		{
			Resp: `{ "names" : "a709c7505efe", "cluster_name" : "docker-cluster", "cluster_uuid" : "8DC8kaEsQ_ClkJtB3OHy9Q", "version" : { "number" : "7.7.1", "build_flavor" : "default", "build_type" : "docker", "build_hash" : "ad56dce891c901a492bb1ee393f12dfff473a423", "build_date" : "2020-05-28T16:30:01.040088Z", "build_snapshot" : false, "lucene_version" : "8.5.1", "minimum_wire_compatibility_version" : "6.8.0", "minimum_index_compatibility_version" : "6.0.0-beta1" }, "tagline" : "You Know, for Search"}`,
			Path: "/",
			Method: []string {"GET", "HEAD"},
			HttpCode: 200,
		},
		{
			Resp: `{"error":{"root_cause":[{"type":"index_not_found_exception","reason":"no such index [t]","resource.type":"index_or_alias","resource.id":"t","index_uuid":"_na_","index":"t"}],"type":"index_not_found_exception","reason":"no such index [t]","resource.type":"index_or_alias","resource.id":"t","index_uuid":"_na_","index":"t"},"status":404}`,
			Path: "/t",
			Method: []string {"GET", "HEAD"},
			HttpCode: 404,

		},
		{
			Resp: `{"weibo":{"mappings":{"properties":{"tel":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"wb_id":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}}},"settings":{"index":{"number_of_shards":"1","provided_name":"weibo","max_result_window":"10000","creation_date":"1610209867377","number_of_replicas":"1","uuid":"EEB86YAhQ-mXs5xb3oN-Jg","version":{"created":"7070199"}}}}}`,
			Path: "/weibo",
			Method: []string {"GET", "HEAD"},
			HttpCode: 200,
		},
		{
			Resp: `{"tianya":{"aliases":{"a":{}},"mappings":{"properties":{"mail":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"name":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"pass":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"tags":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}}},"settings":{"index":{"number_of_shards":"1","provided_name":"tianya","max_result_window":"10000","creation_date":"1614246655737","number_of_replicas":"1","uuid":"AjVKnCUlRaCXKJkngb-MIQ","version":{"created":"7070199"}}}},"weibo":{"aliases":{"a":{}},"mappings":{"properties":{"tel":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"wb_id":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}}},"settings":{"index":{"number_of_shards":"1","provided_name":"weibo","max_result_window":"10000","creation_date":"1610209867377","number_of_replicas":"1","uuid":"EEB86YAhQ-mXs5xb3oN-Jg","version":{"created":"7070199"}}}}}`,
			Path: "/a",
			Method: []string {"GET", "HEAD"},
			HttpCode: 200,
		},
		{
			Resp: `{"weibo_a":{"aliases":{"weibo":{}},"mappings":{"properties":{"tel":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"wb_id":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}}},"settings":{"index":{"number_of_shards":"1","provided_name":"weibo_a","max_result_window":"10000","creation_date":"1610209867377","number_of_replicas":"1","uuid":"EEB86YAhQ-mXs5xb3oN-Jg","version":{"created":"7070199"}}}}}`,
			Path: "/weibo_a",
			Method: []string {"GET", "HEAD"},
			HttpCode: 200,
		},
		{
			Resp: `{"weibo_b":{"aliases":{"weibo":{}},"mappings":{"properties":{"tel":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}},"wb_id":{"type":"text","fields":{"keyword":{"type":"keyword","ignore_above":256}}}}},"settings":{"index":{"number_of_shards":"1","provided_name":"weibo_b","max_result_window":"10000","creation_date":"1610209867377","number_of_replicas":"1","uuid":"EEB86YAhQ-mXs5xb3oN-Jg","version":{"created":"7070199"}}}}}`,
			Path: "/weibo_b",
			Method: []string {"GET", "HEAD"},
			HttpCode: 200,
		},

	})



	es := ES{}
	config := EsConfig{Addresses: ts.URL}
	_,err := es.NewEsObj(config)
	assert.Nil(err)
	//GetIndexStatus

	//未创建索引

	state,err :=es.GetIndexStatus("t")
	assert.Nil(err)
	assert.Equal( IndexStatus{IndexName: "", AliaseName: "", PlanIndexA: false, PlanIndexB: false}, state)


	//多别名
	state, err = es.GetIndexStatus("a")
	assert.Equal(fmt.Errorf("[Index-%s]:[%s]", "a", errno.SysAliasExceedLimit), err )

	//没别名有索引
	state, err = es.GetIndexStatus("weibo")
	assert.Nil(err)
	assert.Equal(IndexStatus{IndexName: "weibo", AliaseName: "", PlanIndexA: false, PlanIndexB: false},state)

	//别名A
	state, err = es.GetIndexStatus("weibo_a")
	assert.Nil(err)
	assert.Equal(IndexStatus{IndexName: "weibo_a", AliaseName: "weibo", PlanIndexA: true, PlanIndexB: false}, state)

	//别名B

	state, err = es.GetIndexStatus("weibo_b")
	assert.Nil(err)
	assert.Equal(IndexStatus{IndexName: "weibo_b", AliaseName: "weibo", PlanIndexA: false, PlanIndexB: true}, state)

}


func mockRouteHandler(m []mockHander) func(http.ResponseWriter,  *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println(r.URL.Path)
		//fmt.Println(r.Method)
		for _,v := range m {
			if ( r.URL.Path == v.Path) {
				_,b := SliceIn(v.Method, r.Method)
				if b {
					w.WriteHeader(v.HttpCode)
					w.Write([]byte(v.Resp))
					return
				}
			}
		}
	}
}

