package model

import (
	"GoWebFrame/orm/errs"
	"errors"
	"reflect"
	"strings"
	"sync"
)

type Register interface {
	// Get 查询元数据
	Get(val any) (*Model, error)
	// Register 注册一个模型
	Register(val any, opts ...Options) (*Model, error)
}

type register struct {
	models sync.Map
}

func NewRegister() Register {
	return &register{}
}

func (r *register) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.models.Load(typ)
	if ok {
		return m.(*Model), nil
	}

	return r.Register(val)
}

func (r *register) Register(val any, opts ...Options) (*Model, error) {

	// 解析model进行注册
	m, err := r.parseModel(val)
	if err != nil {
		return nil, err
	}

	// 利用option模式给model赋值
	for _, opt := range opts {
		err = opt(m)
		if err != nil {
			return nil, err
		}
	}

	// 获取type加入缓存
	typ := reflect.TypeOf(val)
	r.models.Store(typ, m)
	return m, nil
}

func (r *register) parseModel(entity any) (*Model, error) {
	typ := reflect.TypeOf(entity)

	// 只能传入一级指针
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return nil, errors.New("只支持指向结构体的一级指针")
	}

	typ = typ.Elem()

	numField := typ.NumField()
	fieldMap := make(map[string]*Field, numField)
	fields := make([]*Field, numField)
	columnMap := make(map[string]*Field, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)

		// 解析tag
		tags, err := r.parseTag(fd.Tag)
		if err != nil {
			return nil, err
		}
		colName := tags[tagColumn]
		// 拿不到tag则默认获取给他一个
		if colName == "" {
			colName = underscoreName(fd.Name)
		}

		meta := &Field{
			ColName: colName,
			GoName:  fd.Name,
			Typ:     fd.Type,
			Offset:  fd.Offset, // 获取偏移量
			Index:   i,
		}

		// golang中字段map
		fieldMap[fd.Name] = meta
		fields[i] = meta
		// 数据库字段映射字段
		columnMap[colName] = meta
	}

	// 判断是否实现了此方法
	var tableName string
	if fn, ok := entity.(TableName); ok {
		tableName = fn.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}

	return &Model{
		TableName: tableName,
		FieldMap:  fieldMap,
		Fields:    fields,
		ColumnMap: columnMap,
	}, nil

}

// parseTag 解析tag
func (r *register) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get("orm")
	if ormTag == "" {
		// 返回一个空的 map，这样调用者就不需要判断 nil 了
		return map[string]string{}, nil
	}

	res := make(map[string]string)

	// 切分
	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		res[kv[0]] = kv[1]
	}
	return res, nil
}
