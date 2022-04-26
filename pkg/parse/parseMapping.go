package parse

import (
	"fmt"
	"main/configs"
	"main/pkg/errno"
	"strconv"
)

//https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping-types.html
type TypeMappingObj struct {
	IsID bool
	Mold string
}


func TypeMapping(column string, columnMapping []configs.Column) ( TypeMappingObj, error) {

	var typeMapping = map[string]string{}

	typeMapping["id"]   =  "int32"
	typeMapping["integer"] = "int32"
	typeMapping["text"] = "string"
	typeMapping["date"] = "string"
	typeMapping["keyword"] = "string"
	typeMapping["long"] = "int32"

	for _,v := range columnMapping {

		if (v.Name == column) {

			 isID := false

			if column == "id" {
				isID = true
			}

			return TypeMappingObj{
				IsID: isID,
				Mold: typeMapping[v.Type],
			}, nil
			
		}

	}
	//SysTypeUndefined
	return TypeMappingObj{}, fmt.Errorf("[%s]:%s", column,errno.SysTypeUndefined)
}

func StrConversion(types string, value string) (interface{}, error) {

	switch types {
		case "int32":
			v,_ := strconv.Atoi(value)
			return  v , nil
		case "long":
			v,_ := strconv.Atoi(value)
			return  v , nil
		case "string":
			return  value , nil
	}

	return "", fmt.Errorf("[%s]:%s", value, errno.SysTypeUndefined)
}