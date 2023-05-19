package service

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/zhangjunjie6b/mysql2elasticsearch/configs"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/dao"
	"github.com/zhangjunjie6b/mysql2elasticsearch/internal/pkg"
	"gorm.io/driver/mysql"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

var d dao.Dao
var config configs.Content

type mockHander struct {
	Resp     string
	Path     string
	Method   []string
	HttpCode int
}

func TestBulk(t *testing.T) {
	assert := assert.New(t)
	b := Bulk{dao: d}
	mysqlmock := newMockDatabase()

	//mock es
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	//test NewEsObj
	handler = mockRouteHandler([]mockHander{
		{
			Resp:     `{"took":76,"errors":false,"items":[{"index":{"_index":"t","_type":"_doc","_id":"2","_version":2,"result":"updated","_shards":{"total":2,"successful":1,"failed":0},"_seq_no":1,"_primary_term":1,"status":200}}]}`,
			Path:     "/_bulk",
			Method:   []string{"POST"},
			HttpCode: 200,
		},
	})
	es := pkg.ES{}
	esconfig := pkg.EsConfig{Addresses: ts.URL}

	_, err := es.NewEsObj(esconfig)
	assert.Nil(err)

	err = b.Init(config, d, es)
	assert.Nil(err)

	b.config.Reader.Parameter.Connection.BoundarySql = "SELECT min(id) as min,max(id) as max FROM t"
	b.config.Writer.Parameter.BatchSize = 10
	b.config.Reader.Parameter.Connection.QuerySql = "SELECT id,title,keyword FROM t where id >= ? and id <= ?"
	b.config.Writer.Parameter.Column = []configs.Column{
		{Name: "id", Type: "id"},
		{Name: "title", Type: "text"},
		{Name: "keyword", Type: "text"},
		{Name: "vector", Type: "vector"},
	}

	mysqlmock.ExpectQuery("SELECT min(id) as min,max(id) as max FROM t").
		WillReturnRows(sqlmock.NewRows([]string{"min", "max"}).AddRow(1, 10000))

	sections := b.dao.SelectMaxAndMin(b.config.Reader.Parameter.Connection.BoundarySql)
	section := b.Generate(sections, 8)

	expected := map[int]dao.Section{
		1: {Min: 1, Max: 1251},
		2: {Min: 1251, Max: 2501},
		3: {Min: 2501, Max: 3751},
		4: {Min: 3751, Max: 5001},
		5: {Min: 5001, Max: 6251},
		6: {Min: 6251, Max: 7501},
		7: {Min: 7501, Max: 8751},
		8: {Min: 8751, Max: 10001},
	}
	assert.Equal(expected, section)

	for i := expected[1].Min; i <= expected[1].Max; i = i + b.config.Writer.Parameter.BatchSize {

		maxid := i + b.config.Writer.Parameter.BatchSize
		if maxid > expected[1].Max {
			maxid = expected[1].Max
		}
		sql := "SELECT id,title,keyword,vector FROM t where id >= " + strconv.Itoa(i) + " and id <= " +
			strconv.Itoa(maxid)
		expectSql := mysqlmock.ExpectQuery(sql)
		id := i
		rows := sqlmock.NewRows([]string{"id", "title", "keyword"})
		for ; id <= i+b.config.Writer.Parameter.BatchSize; id++ {
			rows.AddRow(id, "title"+strconv.Itoa(id), "keyword"+strconv.Itoa(id))
		}
		expectSql.WillReturnRows(rows)
	}

	b.Run(map[int]dao.Section{1: {Min: 1, Max: 1251}}, "t")

	//todo 单测覆盖不完整，后续考虑用chan处理错误信息和完成状态

}

func newMockDatabase() sqlmock.Sqlmock {

	sqlDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
	)

	if err != nil {
		log.Fatalf("[sqlmock new] %s", err)
	}

	dialector := mysql.New(mysql.Config{
		Conn:                      sqlDB,
		DriverName:                "mysql",
		SkipInitializeWithVersion: true,
	})

	err = d.NewDao(dialector)

	if err != nil {
		log.Fatalf("[gorm open] %s", err)
	}

	return mock
}

func mockRouteHandler(m []mockHander) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		/*	fmt.Println(r.URL.Path)
			fmt.Println(r.Method)
			c := []byte{}
			c,_ = ioutil.ReadAll(r.Body)
			fmt.Println(string(c))*/
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
