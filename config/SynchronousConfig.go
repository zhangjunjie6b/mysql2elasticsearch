package config

import (
	"fmt"
	"github.com/spf13/viper"
	"main/service"
)

type SynchronousConfig struct {
	Job Job
}

type Job struct {
	Setting Setting
	Content Content
}
type Setting struct {
	Speed Speed
}
type Speed struct {
	Channel int
}

type Content struct {
	Reader Reader
	Writer Writer
}


//----- Reader ------
type Reader struct {
	Name string
	Parameter ReaderParameter
}

type ReaderParameter struct {
	Username string
	Password string
	Host	 string
	DbName	 string
	Connection Connection
}

type Connection struct {
	QuerySql string
	JdbcUrl  string
	BoundarySql   string
}
//----- Reader ------


//----- writer ------

type Writer struct {
	Name string
	Parameter  WriterParameter
}

type WriterParameter struct {
	Endpoint string
	AccessId string
	AccessKey string
	Index string
	Types string
	BatchSize int
	Splitter string
	Column []Column
	Dsl string
}

type Column struct {
	Name string
	Type string
	Analyzer string
	Search_analyzer string
	Format string
	Array string
}

//----- writer ------



func NewSynchronousConfig(configName string)  SynchronousConfig {
	config := viper.New()
	config.SetConfigName(configName)
	config.SetConfigType("json")
	config.AddConfigPath("./config")

	err := config.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	synchronousConfig := SynchronousConfig{}
	config.Unmarshal(&synchronousConfig)

	return synchronousConfig
}

func GetMysqlConfig(job Job) string {
	return job.Content.Reader.Parameter.Username + ":" + job.Content.Reader.Parameter.Password  + "@tcp(" + job.Content.Reader.Parameter.Host+ ")/" + job.Content.Reader.Parameter.DbName
}

func GetEsConfig(job Job)  service.EsConfig {
	return service.EsConfig{
		Addresses:             job.Content.Writer.Parameter.Endpoint,
		Username:              job.Content.Writer.Parameter.AccessId,
		Password:              job.Content.Writer.Parameter.AccessKey,
	}
}


/**
	根据顶层配置文件 配置名称 返回 详细的配置文件信息
 */
func JobNameGetSynchronousConfig(jobName string) (SynchronousConfig, bool) {
	configFile := NewConfig()
	synchronousConfig := SynchronousConfig{}
	for _,v := range configFile.JobList {
		if v.Name == jobName {
			synchronousConfig = NewSynchronousConfig(v.FilePath)
			return synchronousConfig, true
		}
	}

	return synchronousConfig, false
}

/**
	根据顶层配置文件  配置名称 返回 es 配置信息
 */
func JobNameGetESConfig(jobName string) (service.EsConfig, SynchronousConfig , bool) {

	synchronousConfig, status := JobNameGetSynchronousConfig(jobName)

	if status == false {
		return service.EsConfig{}, synchronousConfig , false
	}

	return GetEsConfig(synchronousConfig.Job), synchronousConfig, true
}

