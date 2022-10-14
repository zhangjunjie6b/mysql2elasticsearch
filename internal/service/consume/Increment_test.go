package consume

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/olivere/elastic/v7"
	"github.com/smartystreets/goconvey/convey"
	"github.com/zhangjunjie6b/mysql2elasticsearch/configs"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/dao"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/mode"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/pkg"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var D dao.Dao

type mockHander struct {
	Resp     string
	Path     string
	Method   []string
	HttpCode int
}

func TestIncrement_Init(t *testing.T) {
	increment := Increment{}
	increment.Init()
}

func TestIncrement_Handle(t *testing.T) {
	convey.Convey("TestIncrement_Handle", t, func() {
		increment := &Increment{}
		updateF := gomonkey.ApplyPrivateMethod(reflect.TypeOf(increment), "update", func(incr *Increment, data mode.Jobs) error {
			return errors.New("update")
		})

		defer updateF.Reset()

		delF := gomonkey.ApplyPrivateMethod(reflect.TypeOf(increment), "del", func(incr *Increment, data mode.Jobs) error {
			return errors.New("del")
		})
		defer delF.Reset()

		tt := []struct {
			name   string
			jobs   mode.Jobs
			expect error
		}{
			{"update", mode.Jobs{ID: 1, Payload: mode.Payloads{Type: "update"}}, errors.New("update")},
			{"add", mode.Jobs{ID: 2, Payload: mode.Payloads{Type: "add"}}, errors.New("update")},
			{"del", mode.Jobs{ID: 3, Payload: mode.Payloads{Type: "del"}}, errors.New("del")},
			{"default", mode.Jobs{ID: 4, Payload: mode.Payloads{Type: "default"}}, errors.New("payload type not find")},
		}

		for _, tc := range tt {
			convey.Convey(tc.name, func() {
				err := increment.Handle(tc.jobs)
				convey.So(err, convey.ShouldResemble, tc.expect)
			})
		}
	})
}

func TestIncrement_GetConfigInstance(t *testing.T) {
	increment := &Increment{}
	increment.Init()

	convey.Convey("TestIncrement_GetConfigInstance", t, func() {

		tt := []struct {
			name       string
			reEsconfig pkg.EsConfig
			reConfig   configs.SynchronousConfig
			reB        bool
		}{
			{"t1-succeed", pkg.EsConfig{Addresses: "192.0.0.1", Username: "user", Password: "user1"}, configs.SynchronousConfig{}, true},
			{"t2-succeed", pkg.EsConfig{Addresses: "192.0.0.2", Username: "user", Password: "user2"}, configs.SynchronousConfig{}, true},
			{"t3-fail", pkg.EsConfig{}, configs.SynchronousConfig{}, false},
		}

		for _, tc := range tt {

			getESF := gomonkey.ApplyFunc(configs.JobNameGetESConfig, func(name string) (pkg.EsConfig, configs.SynchronousConfig, bool) {
				return tc.reEsconfig, tc.reConfig, tc.reB
			})
			defer getESF.Reset()

			convey.Convey(tc.name, func() {
				sync, es, boo := increment.getConfigInstance(tc.name)
				convey.So(es, convey.ShouldResemble, tc.reEsconfig)
				convey.So(sync, convey.ShouldResemble, tc.reConfig)
				convey.So(boo, convey.ShouldEqual, tc.reB)
			})

		}

		convey.Convey("repeat", func() {
			for _, tc := range tt {
				convey.Convey(tc.name, func() {
					sync, es, boo := increment.getConfigInstance(tc.name)
					convey.So(es, convey.ShouldResemble, tc.reEsconfig)
					convey.So(sync, convey.ShouldResemble, tc.reConfig)
					convey.So(boo, convey.ShouldEqual, tc.reB)
				})
			}
		})

	})
}

func TestIncrement_GetDBInstance(t *testing.T) {
	increment := &Increment{}
	increment.Init()

	convey.Convey("TestIncrement_getDBInstance", t, func() {

		tt := []struct {
			name     string
			reConfig configs.SynchronousConfig
			rebool   bool
			expect   error
		}{

			{"d1-succeed", configs.SynchronousConfig{Job: configs.Job{
				Content: configs.Content{
					Reader: configs.Reader{
						Parameter: configs.ReaderParameter{
							Connection: configs.Connection{
								JdbcUrl:   "u:p@tcp(127.0.0.1)/db_name",
								Increment: "SELECT-1",
							},
						},
					},
				},
			}}, true, nil},
			{"d2-succeed", configs.SynchronousConfig{Job: configs.Job{
				Content: configs.Content{
					Reader: configs.Reader{
						Parameter: configs.ReaderParameter{
							Connection: configs.Connection{
								JdbcUrl:   "u:p@tcp(127.0.0.2)/db_name",
								Increment: "SELECT-2",
							},
						},
					},
				},
			}}, true, nil},
			{"d3-fail", configs.SynchronousConfig{}, false, ErrNoConfigFile},
		}

		for _, v := range tt {

			convey.Convey(v.name, func() {

				openF := gomonkey.ApplyFunc(mysql.Open, func(dsn string) gorm.Dialector {

					convey.Convey("dao.NewDao", func() {
						convey.So(dsn, convey.ShouldEqual, v.reConfig.Job.Content.Reader.Parameter.Connection.JdbcUrl)
					})

					sqlDB, _, _ := sqlmock.New(
						sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
					)

					config := mysql.Config{
						Conn:                      sqlDB,
						DriverName:                "mysql",
						SkipInitializeWithVersion: true,
					}
					return mysql.New(config)
				})
				defer openF.Reset()

				getConfigInstanceF := gomonkey.ApplyPrivateMethod(reflect.TypeOf(increment), "getConfigInstance",
					func(jobname string) (configs.SynchronousConfig, pkg.EsConfig, bool) {
						return v.reConfig, pkg.EsConfig{}, v.rebool
					})
				defer getConfigInstanceF.Reset()

				convey.Convey("getDBInstance return", func() {
					db, err := increment.getDBInstance(v.name)
					convey.So(db.sql, convey.ShouldEqual, v.reConfig.Job.Content.Reader.Parameter.Connection.Increment)
					convey.So(err, convey.ShouldResemble, v.expect)
				})

			})

		}

		convey.Convey("repeat", func() {
			for _, v := range tt {
				if v.rebool {
					convey.Convey(v.name, func() {
						db, err := increment.getDBInstance(v.name)
						convey.Convey("getDBInstance return", func() {
							convey.So(err, convey.ShouldResemble, v.expect)
							convey.So(db.sql, convey.ShouldEqual, v.reConfig.Job.Content.Reader.Parameter.Connection.Increment)
						})
					})
				}
			}
		})

	})

}

func TestIncrement_GetEsInstance(t *testing.T) {
	increment := &Increment{}
	increment.Init()

	//mock
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()
	//test NewEsObj
	handler = mockRouteHandler([]mockHander{
		{
			Resp:     `{ "names" : "a709c7505efe", "cluster_name" : "docker-cluster", "cluster_uuid" : "8DC8kaEsQ_ClkJtB3OHy9Q", "version" : { "number" : "7.7.1", "build_flavor" : "default", "build_type" : "docker", "build_hash" : "ad56dce891c901a492bb1ee393f12dfff473a423", "build_date" : "2020-05-28T16:30:01.040088Z", "build_snapshot" : false, "lucene_version" : "8.5.1", "minimum_wire_compatibility_version" : "6.8.0", "minimum_index_compatibility_version" : "6.0.0-beta1" }, "tagline" : "You Know, for Search"}`,
			Path:     "/",
			Method:   []string{"GET", "HEAD"},
			HttpCode: 200,
		},
	})

	convey.Convey("TestIncrement_getEsInstance", t, func() {
		tt := []struct {
			name        string
			reesConfig  pkg.EsConfig
			rebool      bool
			expecststus bool
			expectErr   error
		}{
			{"e1-succeed", pkg.EsConfig{Addresses: ts.URL}, true, true, nil},
			{"e2-fail", pkg.EsConfig{Addresses: "http://127.0.0.1:9200"}, true, false, elastic.ErrNoClient},
			{"e3-fail", pkg.EsConfig{}, false, false, ErrNoConfigFile},
		}

		for _, v := range tt {

			getConfigInstanceF := gomonkey.ApplyPrivateMethod(reflect.TypeOf(increment), "getConfigInstance",
				func(name string) (configs.SynchronousConfig, pkg.EsConfig, bool) {
					return configs.SynchronousConfig{}, v.reesConfig, v.rebool
				})
			defer getConfigInstanceF.Reset()

			convey.Convey(v.name, func() {
				_, err := increment.getEsInstance(v.name)
				convey.So(errors.Is(err, v.expectErr), convey.ShouldEqual, true)
			})

		}

		convey.Convey("repeat", func() {
			for _, v := range tt {

				if v.expecststus {
					_, err := increment.getEsInstance(v.name)
					convey.Convey(v.name, func() {
						convey.So(errors.Is(err, v.expectErr), convey.ShouldEqual, true)
					})
				}

			}
		})

	})

}

func TestIncrement_Update(t *testing.T) {
	increment := &Increment{}
	increment.Init()

	//mock
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	//test NewEsObj
	handler = mockRouteHandler([]mockHander{
		{
			Resp:     `{ "names" : "a709c7505efe", "cluster_name" : "docker-cluster", "cluster_uuid" : "8DC8kaEsQ_ClkJtB3OHy9Q", "version" : { "number" : "7.7.1", "build_flavor" : "default", "build_type" : "docker", "build_hash" : "ad56dce891c901a492bb1ee393f12dfff473a423", "build_date" : "2020-05-28T16:30:01.040088Z", "build_snapshot" : false, "lucene_version" : "8.5.1", "minimum_wire_compatibility_version" : "6.8.0", "minimum_index_compatibility_version" : "6.0.0-beta1" }, "tagline" : "You Know, for Search"}`,
			Path:     "/",
			Method:   []string{"GET", "HEAD"},
			HttpCode: 200,
		},
		{
			Resp:     `{"_shards":{"total":0,"successful":0,"failed":0},"_index":"test","_type":"_doc","_id":"1","_version":7,"result":"noop"}`,
			Path:     "/update1-succeed/_update/28866",
			Method:   []string{"POST", "HEAD"},
			HttpCode: 200,
		},
		{
			Resp:     `{"status":400}`,
			Path:     "/update3-fail/_update/288669",
			Method:   []string{"POST", "HEAD"},
			HttpCode: 400,
		},
	})
	mock := newMockData()

	convey.Convey("TestIncrement_Update", t, func() {

		tt := []struct {
			name        string
			db          dbObj
			rows        *sqlmock.Rows
			esConfig    pkg.EsConfig
			data        mode.Jobs
			Column      []configs.Column
			expectError interface{}
		}{
			{"update1-succeed", dbObj{sql: "select where id = ?", dao: D},
				sqlmock.NewRows([]string{"id", "title", "keyword"}).
					AddRow(1, "测试1", "测试1"),
				pkg.EsConfig{Addresses: ts.URL},
				mode.Jobs{Payload: mode.Payloads{Id: 28866, EsIndexName: "update1-succeed"}},
				[]configs.Column{{Name: "id", Type: "id"}, {Name: "title", Type: "keyword"}, {Name: "keyword", Type: "keyword"}},
				nil,
			},
			{"update2-fail", dbObj{sql: "select where id = ?", dao: D},
				sqlmock.NewRows([]string{"id", "title", "keyword"}),
				pkg.EsConfig{Addresses: ts.URL},
				mode.Jobs{},
				[]configs.Column{},
				errors.New("doc is null"),
			},
			{"update3-fail", dbObj{sql: "select where id = ?", dao: D},
				sqlmock.NewRows([]string{"id", "title", "keyword"}).
					AddRow(1, "测试1", "测试1"),
				pkg.EsConfig{Addresses: ts.URL},
				mode.Jobs{Payload: mode.Payloads{Id: 288669, EsIndexName: "update3-fail"}},
				[]configs.Column{{Name: "id", Type: "id"}, {Name: "title", Type: "keyword"}, {Name: "keyword", Type: "keyword"}},
				&elastic.Error{Status: 400, Details: nil},
			},
		} //ErrorDetails

		for _, v := range tt {

			mock.ExpectQuery(v.db.sql).WillReturnRows(v.rows)

			db := gomonkey.ApplyPrivateMethod(increment, "getDBInstance", func(jobname string) (dbObj, error) {
				return v.db, nil
			})

			defer db.Reset()

			es := gomonkey.ApplyPrivateMethod(reflect.TypeOf(increment), "getConfigInstance",
				func(name string) (configs.SynchronousConfig, pkg.EsConfig, bool) {
					return configs.SynchronousConfig{}, v.esConfig, true
				})

			defer es.Reset()

			syn := configs.SynchronousConfig{Job: configs.Job{Content: configs.Content{Writer: configs.Writer{Parameter: configs.WriterParameter{Column: v.Column}}}}}
			increment.synchronousConfig[v.name] = syn

			convey.Convey(v.name, func() {
				error := increment.update(v.data)
				convey.So(error, convey.ShouldResemble, v.expectError)
			})

		}

	})

}

func TestIncrement_Del(t *testing.T) {
	increment := &Increment{}
	increment.Init()

	//mock
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	//test NewEsObj
	handler = mockRouteHandler([]mockHander{
		{
			Resp:     `{ "names" : "a709c7505efe", "cluster_name" : "docker-cluster", "cluster_uuid" : "8DC8kaEsQ_ClkJtB3OHy9Q", "version" : { "number" : "7.7.1", "build_flavor" : "default", "build_type" : "docker", "build_hash" : "ad56dce891c901a492bb1ee393f12dfff473a423", "build_date" : "2020-05-28T16:30:01.040088Z", "build_snapshot" : false, "lucene_version" : "8.5.1", "minimum_wire_compatibility_version" : "6.8.0", "minimum_index_compatibility_version" : "6.0.0-beta1" }, "tagline" : "You Know, for Search"}`,
			Path:     "/",
			Method:   []string{"GET", "HEAD"},
			HttpCode: 200,
		},
		{
			Resp:     `{"_index":"t1_a","_type":"_doc","_id":"41768","_version":4,"result":"deleted","_shards":{"total":2,"successful":1,"failed":0},"_seq_no":536,"_primary_term":1}`,
			Path:     "/del1-succeed/_doc/28866",
			Method:   []string{"DELETE", "HEAD"},
			HttpCode: 200,
		},

		{
			Resp:     `{"_index":"t1_a","_type":"_doc","_id":"28866","_version":8,"result":"not_found","_shards":{"total":2,"successful":1,"failed":0},"_seq_no":409,"_primary_term":1}`,
			Path:     "/del2-fild/_doc/28866",
			Method:   []string{"DELETE", "HEAD"},
			HttpCode: 404,
		},
	})

	convey.Convey("TestIncrement_Del", t, func() {

		tt := []struct {
			name        string
			esConfig    pkg.EsConfig
			data        mode.Jobs
			expectError interface{}
		}{
			{"del1-succeed", pkg.EsConfig{Addresses: ts.URL},
				mode.Jobs{Payload: mode.Payloads{Id: 28866, EsIndexName: "del1-succeed"}}, nil,
			},
			{"del2-fild", pkg.EsConfig{Addresses: ts.URL},
				mode.Jobs{Payload: mode.Payloads{Id: 28866, EsIndexName: "del2-fild"}},
				&elastic.Error{Status: 404, Details: nil},
			},
		}

		for _, v := range tt {
			es := gomonkey.ApplyPrivateMethod(reflect.TypeOf(increment), "getConfigInstance",
				func(name string) (configs.SynchronousConfig, pkg.EsConfig, bool) {
					return configs.SynchronousConfig{}, v.esConfig, true
				})

			defer es.Reset()

			convey.Convey(v.name, func() {
				err := increment.del(v.data)
				convey.So(err, convey.ShouldResemble, v.expectError)
			})

		}

	})

}

func newMockData() sqlmock.Sqlmock {

	sqlDB, mock, err := sqlmock.New()

	if err != nil {
		log.Fatalf("[sqlmock new] %s", err)
	}

	config := mysql.Config{
		Conn:                      sqlDB,
		DriverName:                "mysql",
		SkipInitializeWithVersion: true,
	}

	dialector := mysql.New(config)

	err = D.NewDao(dialector)

	if err != nil {
		log.Fatalf("[gorm open] %s", err)
	}

	return mock
}

func mockRouteHandler(m []mockHander) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println(r.URL.Path)
		//fmt.Println(r.Method)
		for _, v := range m {
			if r.URL.Path == v.Path {
				_, b := pkg.SliceIn(v.Method, r.Method)
				if b {
					w.WriteHeader(v.HttpCode)
					w.Write([]byte(v.Resp))
					return
				}
			}
		}
	}
}
