package valuer

import (
	"GoWebFrame/orm/interal/model"
	"database/sql"
	"errors"
	"reflect"
	"unsafe"
)

type UnsafeValues struct {
	t     any
	model *model.Model
}

func NewUnsafeValues(t any, model *model.Model) Valuer {
	return &UnsafeValues{
		t:     t,
		model: model,
	}
}

func (u UnsafeValues) SetColumns(rows *sql.Rows) error {
	if !rows.Next() {
		return sql.ErrNoRows
	}

	// 获取读到的所有列
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	t := u.t
	vals := make([]any, 0, len(cs))
	// 获取地址, unsafe.Pointer 和  uintptr的区别，   unsafe.Pointer： go会在GC之后维护，而uintptr不会
	addr := unsafe.Pointer(reflect.ValueOf(t).Pointer())

	for i, c := range cs {
		fd, ok := u.model.ColumnMap[c]
		if !ok {
			return errors.New("找不到此字段")
		}
		// 反射创建一个实例
		//这里创建的实例要是原本类型的指针
		//离谱 fd.type = int 那么val = *int
		// 这段相当于获取到每个字段的类型以及内存地址
		vals[i] = reflect.NewAt(fd.Typ, unsafe.Pointer(uintptr(addr)+fd.Offset))
	}
	// 直接赋值
	return rows.Scan(vals...)
}
