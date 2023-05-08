package model

import (
	"reflect"
	"unicode"
)

const (
	tagColumn = "column"
)

type Model struct {
	TableName string
	FieldMap  map[string]*Field
	ColumnMap map[string]*Field
}

// TableName 表名接口
// @Description:  实现此接口代表用这个表名替换默认名称
type TableName interface {
	TableName() string
}

type Field struct {
	// 列名
	ColName string
	// 代码结构体名
	GoName string
	// 类型
	Typ reflect.Type
	// 偏移量
	Offset uintptr
}

type Options func(model *Model) error

func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}
	}
	return string(buf)
}
