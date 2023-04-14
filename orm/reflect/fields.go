package reflect

import (
	"errors"
	"reflect"
)

func IterateFields(entity any) (map[string]any, error) {

	if entity == nil {
		return nil, errors.New("不支持传入nil")
	}

	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)

	if val.IsZero() {
		// 预防传入的结构体，但所有的bit位都是0值的情况
		return nil, errors.New("不支持0值")
	}

	// 这里用for不用if是因为可能会存在多级指针， if只能判断一层
	for typ.Kind() == reflect.Ptr {
		// 拿到指针指向的对象
		typ = typ.Elem()
		val = val.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, errors.New("不支持类型")
	}

	res := make(map[string]any, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		// 字段类型
		fieldType := typ.Field(i)
		// 字段值
		fieldVal := val.Field(i)

		// 判断是否为可导出字段，以大写字母开头
		if fieldType.IsExported() {
			res[fieldType.Name] = fieldVal.Interface()
		} else {
			// 私有的拿不到，用0值填充
			res[fieldType.Name] = reflect.Zero(fieldType.Type)
		}

	}

	return res, nil
}

func SetField(entity any, field string, newValue any) error {
	val := reflect.ValueOf(entity)

	for val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}

	name := val.FieldByName(field)
	if !name.CanSet() {
		return errors.New("不可修改字段")
	}

	name.Set(reflect.ValueOf(newValue))
	return nil
}
