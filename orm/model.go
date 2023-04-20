package orm

import (
	"unicode"
)

const (
	tagColumn = "column"
)

type Model struct {
	tableName string
	fields    map[string]*field
}

// TableName 表名接口
// @Description:  实现此接口代表用这个表名替换默认名称
type TableName interface {
	TableName() string
}

type field struct {
	// 列名
	colName string
}

type ModelOpt func(model *Model) error

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
