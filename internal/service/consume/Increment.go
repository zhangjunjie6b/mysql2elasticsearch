package consume

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"main/configs"
	"main/internal/dao"
	"main/internal/mode"
	"main/internal/pkg"
	"strconv"
	"strings"
	"time"
)

var(
	ErrNoConfigFile = errors.New("config file read error")
)

type Increment struct {
	dbObj map[string] dbObj
	esObj map[string] pkg.ES
	synchronousConfig map[string] configs.SynchronousConfig
	esConfig map[string] pkg.EsConfig
}

type dbObj struct {
	dao dao.Dao
	sql string
}

func  (c *Increment) Init() {
	c.dbObj = make(map[string] dbObj)
	c.esObj = make(map[string] pkg.ES)
	c.synchronousConfig = make(map[string] configs.SynchronousConfig)
	c.esConfig = make(map[string] pkg.EsConfig)
}

func (c Increment) Handle(data interface{}) error{

	da,_ := data.(mode.Jobs)
	switch da.Payload.Type {
	case "update": return c.update(da)
	case "add": return c.update(da)
	case "del": return c.del(da)
	default: return errors.New("payload type not find")
	}

	return nil
}

func (c *Increment) getConfigInstance(jobname string) (configs.SynchronousConfig, pkg.EsConfig, bool) {
	synchronousConfig, esbool := c.synchronousConfig[jobname]
	esconfig, dbbool := c.esConfig[jobname]

	var ok bool
	if esbool && dbbool {
		return synchronousConfig, esconfig, true
	} else {
		c.esConfig[jobname], c.synchronousConfig[jobname], ok = configs.JobNameGetESConfig(jobname)
		if ok == false {
			delete(c.esConfig, jobname)
			delete(c.synchronousConfig, jobname)
			return configs.SynchronousConfig{}, pkg.EsConfig{},false
		}
		return c.synchronousConfig[jobname], c.esConfig[jobname], true
	}
}

func (c *Increment) getDBInstance (jobname string) (dbObj, error) {
	value, ok :=  c.dbObj[jobname]
	db := dbObj{}

	if ok {
		return value,nil
	} else {

		synchronousConfig,_,ok := c.getConfigInstance(jobname)

		if ok == false {
			return db, ErrNoConfigFile
		}

		dao := dao.Dao{}

		err := dao.NewDao(mysql.Open(synchronousConfig.Job.Content.Reader.Parameter.Connection.JdbcUrl))

		if err != nil {
			return db, err
		}
		db.dao = dao
		db.sql = synchronousConfig.Job.Content.Reader.Parameter.Connection.Increment

		c.dbObj[jobname] = db

		return db,nil
	}

}

func (c *Increment) getEsInstance (jobname string) (pkg.ES,error) {
	value, ok :=  c.esObj[jobname]
	es := pkg.ES{}

	if ok {
		return value,nil
	} else {

		_,esConfig,ok := c.getConfigInstance(jobname)

		if ok == false {
			return es, ErrNoConfigFile
		}

		_,err := es.NewEsObj(esConfig)

		if err != nil {
			return es, err
		}

		c.esObj[jobname] = es
		return es ,nil
	}
}

func (c *Increment)update(data mode.Jobs) error {

	db,err := c.getDBInstance(data.Payload.EsIndexName)
	if err != nil  {
		return err
	}

	es, err := c.getEsInstance(data.Payload.EsIndexName)
	if err != nil  {
		return err
	}
	defer es.Ctx.Done()

	var rows *sql.Rows

	timesCount := 0

	for true {

		sqlQuery := strings.Replace(db.sql, "?", strconv.Itoa(data.Payload.Id), 1)
		rows,err = db.dao.GetClient().Raw(sqlQuery).Rows()

		if err != nil {
			if timesCount > 10 {
				panic(any(err))
			}
			time.Sleep(time.Second * 10)
			fmt.Printf("query error : %d", timesCount)
			timesCount++
		} else {
			break
		}
	}

	doc,err:= db.dao.ResultTostring(rows,
		c.synchronousConfig[data.Payload.EsIndexName].Job.Content.Writer.Parameter.Column)

	if err != nil  {
		return err
	}


	if len(doc) < 1 {
		return errors.New("doc is null")
	}

	if doc[0].FieldID.Status {

		var v interface{}
		json.Unmarshal([]byte(doc[0].JsonString), &v)

		_,err := es.Client.Update().
			Index(data.Payload.EsIndexName).
			Id(strconv.Itoa(data.Payload.Id)).
			Doc(&v).
			DocAsUpsert(true).
			Do(es.Ctx)

		if err != nil  {
			return err
		}

	} else {
		errors.New("FieldID is null")
	}

	return nil

}

func (c *Increment) del (data mode.Jobs) error {

	_, err := c.getEsInstance(data.Payload.EsIndexName)

	if err != nil  {
		return err
	}
	 _,err = c.esObj[data.Payload.EsIndexName].Client.Delete().
		Index(data.Payload.EsIndexName).
		Id(strconv.Itoa(data.Payload.Id)).
		Do(c.esObj[data.Payload.EsIndexName].Ctx)

	if err != nil  {
		return err
	}
	return nil
}
