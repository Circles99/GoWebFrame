package orm

import (
	"GoWebFrame/orm/errs"
	"GoWebFrame/orm/interal/model"
	"reflect"
)

type OnDuplicateBuilder[T any] struct {
	i *Inserter[T]
}

type OnDuplicateKey struct {
	assigns []Assignable
}

func (o *OnDuplicateBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicate = &OnDuplicateKey{assigns: assigns}

	return o.i
}

type Inserter[T any] struct {
	builder
	// 字段
	columns []string
	// 赋值
	values []*T
	db     *DB
	args   []any
	// 方案二
	onDuplicate *OnDuplicateKey
}

func NewInserter[T any](db *DB) *Inserter[T] {
	return &Inserter[T]{
		db: db,
		builder: builder{
			dialect: db.dialect,
			quoter:  db.dialect.quoter(),
		},
	}
}

func (i *Inserter[T]) Columns(cols ...string) *Inserter[T] {
	i.columns = cols
	return i
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.values = vals
	return i
}

func (i *Inserter[T]) OnDuplicateKey() *OnDuplicateBuilder[T] {
	return &OnDuplicateBuilder[T]{
		i: i,
	}
}

func (i Inserter[T]) Build() (*Query, error) {
	if len(i.values) == 0 {
		return nil, errs.ErrInsertZeroRow
	}

	// 解析model, values都是model
	m, err := i.db.r.Get(i.values[0])
	if err != nil {
		return nil, err
	}
	// 拼接sql
	i.sb.WriteString("INSERT INTO ")
	i.quote(i.model.TableName)
	i.sb.WriteString("(")
	// 对比字段值和传进来的值
	fields := m.Fields

	if len(i.columns) > 0 {
		fields = make([]*model.Field, len(i.columns))
		for id, c := range i.columns {
			field, ok := m.FieldMap[c]
			if !ok {
				return nil, errs.NewErrUnknownField(c)
			}
			fields[id] = field
		}
	}
	// 进行field拼接
	for idx, field := range fields {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(field.ColName)
	}
	i.sb.WriteString(") VALUES")

	// 拼接values
	// 每个val都是一个model
	for idx, value := range i.values {
		if idx > 0 {
			i.sb.WriteByte(',')
		}
		// 反射获取model
		refVal := reflect.ValueOf(value).Elem()
		i.sb.WriteByte('(')
		// 把model里的值都拿出来进行拼接
		for fIdx, field := range fields {
			if fIdx > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			// 根据index获取model的特定字段的值
			fdVal := refVal.Field(field.Index)
			i.args = append(i.args, fdVal.Interface())
		}
		i.sb.WriteByte(')')
	}

	if len(i.onDuplicate.assigns) > 0 {
		i.dialect.buildUpsert(&i.builder, i.onDuplicate)
	}

	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}
