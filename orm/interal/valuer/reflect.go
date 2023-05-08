package valuer

import (
	"GoWebFrame/orm/interal/model"
	"database/sql"
	"errors"
	"reflect"
)

type ReflectValues struct {
	t     any
	model *model.Model
}

func NewReflectValues(t any, model *model.Model) Valuer {
	return &ReflectValues{
		t:     t,
		model: model,
	}
}

func (r ReflectValues) SetColumns(rows *sql.Rows) error {
	if !rows.Next() {
		return sql.ErrNoRows
	}

	// 获取读到的所有列
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))

	// 这一步相当于准备箱子
	for i, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errors.New("找不到此字段")
		}
		// 反射创建一个实例
		//这里创建的实例要是原本类型的指针
		val := reflect.New(fd.Typ)
		// 因为 Scan 要指针，所以我们在这里，不需要调用 Elem， interface出来就是ptr
		vals[i] = val.Interface()

		// fd.Type 是 int，那么  reflect.New(fd.typ) 是 *int
		//要调用ele。 因为fd.type = int, val是*int, 调用elem， val变成int
		//valElems[i] 等同于 reflect.ValueOf(vals[i]).Elem()
		valElems[i] = val.Elem()
	}

	// 类型要匹配
	// 顺序要匹配
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}

	tp := r.t
	// 赋值进model中
	tpValueElem := reflect.ValueOf(tp).Elem()
	// 这一步相当于把箱子搬到对应位置
	for i, c := range cs {
		fd, ok := r.model.ColumnMap[c]
		if !ok {
			return errors.New("找不到此字段")
		}
		// 都使用指针进行赋值

		tpValueElem.FieldByName(fd.GoName).Set(valElems[i])

	}
	return nil
}
