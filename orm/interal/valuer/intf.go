package valuer

import (
	"GoWebFrame/orm/interal/model"
	"database/sql"
)

// Valuer 是对结构体实例的内部抽象
type Valuer interface {
	// SetColumns 设置新值
	SetColumns(rows *sql.Rows) error
}

// Creator 本质上也可以看所是 factory 模式，极其简单的 factory 模式
type Creator func(t any, model *model.Model) Valuer
