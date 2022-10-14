package parse

import (
	"github.com/stretchr/testify/assert"
	"github.com/zhangjunjie6b/mysql2elasticsearch/configs"
	"testing"
)

func TestTypeMapping(t *testing.T) {
	type list struct {
		Column    []configs.Column
		TestValue string
		Assert    TypeMappingObj
	}

	configColumn := []configs.Column{
		{Name: "id", Type: "id"},
		{Name: "pic_id", Type: "integer"},
		{Name: "keyword", Type: "text"},
		{Name: "name", Type: "keyword"},
		{Name: "uid", Type: "long"},
		{Name: "id-2", Type: "id"},
	}

	l := []list{
		{
			Column:    configColumn,
			TestValue: "id",
			Assert:    TypeMappingObj{true, "int32"},
		},
		{
			Column:    configColumn,
			TestValue: "pic_id",
			Assert:    TypeMappingObj{false, "int32"},
		},
		{
			Column:    configColumn,
			TestValue: "keyword",
			Assert:    TypeMappingObj{false, "string"},
		},
		{
			Column:    configColumn,
			TestValue: "name",
			Assert:    TypeMappingObj{false, "string"},
		},
		{
			Column:    configColumn,
			TestValue: "uid",
			Assert:    TypeMappingObj{false, "int32"},
		},
		{
			Column:    configColumn,
			TestValue: "id-2",
			Assert:    TypeMappingObj{true, "int32"},
		},
	}

	assert := assert.New(t)
	for _, v := range l {
		mapping, _ := TypeMapping(v.TestValue, v.Column)
		assert.Equal(v.Assert, mapping, v.TestValue)
	}

	_, err := TypeMapping("err", []configs.Column{{Name: "err", Type: "err"}})
	assert.Error(err, "column undefined")
}

func TestStrConversion(t *testing.T) {
	value, err := StrConversion("int32", "18")
	assert.NoError(t, err)
	assert.Equal(t, value, 18)

	value, err = StrConversion("long", "18")
	assert.NoError(t, err)
	assert.Equal(t, value, 18)

	value, err = StrConversion("string", "18")
	assert.NoError(t, err)
	assert.Equal(t, value, "18")

	value, err = StrConversion("bool", "true")
	assert.Error(t, err)
	assert.Empty(t, value, value)

}
