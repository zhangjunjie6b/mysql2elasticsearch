# 简介
 虽然 logstash dts 等工具提供了很好的多数据源同步订阅方案，但是实际项目中并不会把服务器的上机权限或云厂商账号开放给所有成员。
 这个小工具算对这类场景的一种补充,它可以帮助我们从mysql中取出数据并推送到elasticsearch上

![16291791332153](http://pic.phpzjj.com/mweb/2021/08/17/16291791332153.jpg)

# 特点

 * 基于 elasticsearch  REST APIs 理论无版本上的兼容问题
 * 配置简单仅需3步即可完成 1. 配置 mapping  2. 编写获取源数据sql语句 3. 对应字段映射关系
 * 提供WebGUI配置完成后日常维护无需上机

# 流程

![img.png](https://pic.phpzjj.com/go/image/2021/9/9/a83e2ecc-a3da-4499-bd69-fa712c068e50.jpeg)


# 更新历史

版本|更新内容
:-:|:-:
v1.0|实现基础功能
v2.0|调整项目目录结构&&修复进度条准确性



# 安装

Docker

```
docker run -p 9102:9102  -v /config/:/root/bin/config  registry.cn-shanghai.aliyuncs.com/zhangjunjie6b/mysql2elasticsearch:[镜像版本号]
```

 下载对应平台运行，或源码编译安装

MAC 

```
# Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
 
# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
```

Linux
```
# Mac
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build
 
# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build
```

Windows
```
# Mac
SET CGO_ENABLED=0
SET GOOS=darwin
SET GOARCH=amd64
go build 
 
# Linux
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build 

```

# 配置

```
config.json //需要推送的索引列表和配置文件对应关系

{
  "jobList" : [
    {"name": "本地" , "filePath" : "MysqlReader.json"},
    {"name": "aaa" , "filePath" : "T1.json"}
  ]
}

-------------
filePath.json //单个配置文件

{
  "job": {
    "setting": {
      "speed": {
        "channel" : 8 //worker 数量
      }
    },
    "content": {
        "reader": {
          "name": "mysqlreader",
          "parameter": {
            "connection": {
                "jdbcUrl" : "账号:密码@tcp(ip)/库",
                "querySql": "获取源数据sql",
                "boundarySql" : "SELECT min(id) as min,max(id) as max FROM t"
            }
          }
        },
        "writer": {
          "name": "elasticsearchwriter",
          "parameter": {
            "endpoint": "http://ip:9200",
            "accessId": "账号",
            "accessKey": "密码",
            "index": "test", //索引名称
            "type": "_doc",
            "batchSize": 10000,//一次bulk的数量
            "splitter": ",",
            "column" : [
              {"name": "id", "type": "id"},
              {"name": "pic_id", "type": "text"},
              {"name": "dl_num", "type": "integer"}
            ],
            "dsl" : "{\n  \"settings\": {\n    \"index\": {\n      \"sort.field\": \"pr\",\n      \"sort.order\": \"desc\",\n      \"store.type\": \"hybridfs\",\n      \"number_of_shards\": 1, \n      \"number_of_replicas\": 1,\n      \"similarity\" : {\n          \"default\" : {\n            \"type\" : \"BM25\",\n            \"b\": 0,\n            \"k1\": 1.2\n          }\n      }\n    }\n  },\n  \"mappings\": {\n    \"properties\": {\n      \"pic_id\": {\n        \"type\": \"keyword\"\n      },\n      \"author\": {\n        \"type\": \"keyword\"\n      },\n      \"title\": {\n        \"type\": \"text\",\n        \"analyzer\": \"ik_max_word\",\n        \"search_analyzer\": \"ik_smart\"\n      },\n      \"title_smart\": {\n        \"type\": \"text\",\n        \"analyzer\": \"ik_smart\"\n      },\n      \"keyword\": {\n        \"type\": \"text\",\n        \"analyzer\": \"ik_smart\"\n      },\n      \"source\": {\n        \"type\": \"keyword\"\n      },\n      \"tags\": {\n        \"type\": \"text\",\n        \"analyzer\": \"ik_smart\"\n      },\n      \"c1\": {\n        \"type\": \"keyword\"\n      },\n      \"dl_num\": {\n        \"type\": \"integer\",\n        \"doc_values\": true\n      },\n      \"cl_num\": {\n        \"type\": \"integer\"\n      },\n      \"format\": {\n        \"type\": \"keyword\"\n      },\n      \"picwidth\": {\n        \"type\": \"integer\"\n      },\n      \"picheight\": {\n        \"type\": \"integer\"\n      },\n      \"format_type\": {\n        \"type\": \"keyword\"\n      },\n      \"filesize\": {\n        \"type\": \"integer\"\n      },\n      \"pr\": {\n        \"type\": \"integer\",\n        \"doc_values\": true\n      },\n      \"down\": {\n        \"type\": \"integer\",\n        \"doc_values\": true\n      },\n      \"t_id\": {\n        \"type\": \"text\"\n      },\n      \"create_time\": {\n        \"type\": \"integer\",\n        \"doc_values\": true\n      },\n      \"hide_sort_by_type\": {\n        \"type\": \"keyword\"\n      },\n      \"picurl\": {\n        \"type\": \"keyword\"\n      }\n    }\n  }\n}" //mapping 语句
          }
        }
      }

  }
}





```
