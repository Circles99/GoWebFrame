package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	sb    *strings.Builder // sb在指针中也会引起复制，所以需要获取指针
	tbl   string
	model *Model
	where []Predicate
	args  []any
	db    *DB
}

// 应该在DB上创建Selector比较合适 golang 不支持这种写法 Method cannot have type parameters
//func (d DB) NewSelector[T any]() *Selector[T] {
//	return &Selector[T]{
//		db: d,
//	}
//}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		db: db,
	}
}

// From 加入表名，为了链式调用返回Selector[T]
func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.tbl = tbl
	return s
}

func (s *Selector[T]) Where(p ...Predicate) *Selector[T] {
	s.where = append(s.where, p...)
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)

	s.sb = &strings.Builder{}
	s.model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}

	s.sb.WriteString("SELECT * FROM ")

	if s.tbl == "" {
		//var t T
		//// 获取类型
		//typ := reflect.TypeOf(t)
		// 获取名字
		s.sb.WriteString("`")
		//s.sb.WriteString(typ.Name())
		s.sb.WriteString(s.model.tableName)
		s.sb.WriteString("`")
	} else {
		s.sb.WriteString(s.tbl)
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		// 拿出第0个, 下面循环从1开始，进行and, 二叉树
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			// 每次返回一个新的predicate,重新赋值
			p = p.And(s.where[1])
		}
		err := s.buildExpression(p)
		if err != nil {
			return nil, err
		}
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	query, err := s.Build()
	if err != nil {
		return nil, err
	}
	// 获取数据
	rows, err := s.db.db.QueryContext(ctx, query.SQL, query.Args)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	// 获取读到的所有列
	cs, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	tp := new(T)
	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))

	// 这一步相当于准备箱子
	for _, c := range cs {
		fd, ok := s.model.ColumnMap[c]
		if !ok {
			return nil, errors.New("找不到此字段")
		}
		// 反射创建一个实例
		//这里创建的实例要是原本类型的指针
		//离谱 fd.type = int 那么val = *int
		val := reflect.New(fd.Typ)
		vals = append(vals, val.Interface())
		//要调用ele。 因为fd.type = int, val是*int
		valElems = append(valElems, val.Elem())
	}

	// 类型要匹配
	// 顺序要匹配
	err = rows.Scan(vals...)
	if err != nil {
		return nil, err
	}

	// 赋值进model中
	tpValueElem := reflect.ValueOf(tp).Elem()
	// 这一步相当于把箱子搬到对应位置
	for i, c := range cs {
		fd, ok := s.model.ColumnMap[c]
		if !ok {
			return nil, errors.New("找不到此字段")
		}
		// 都使用指针
		tpValueElem.FieldByName(fd.goName).Set(valElems[i])

	}

	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

// buildExpression
// Expression 是从 sql where 中获取的字段， model中的field 从 struct中获取
func (s *Selector[T]) buildExpression(e Expression) error {
	switch expr := e.(type) {
	case nil:
		return nil
	case Predicate:
		// 进行左右两边递归调用拼接值
		_, ok := expr.left.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.left); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}
		// 拼接中间
		s.sb.WriteByte(' ')
		s.sb.WriteString(expr.op.String())
		s.sb.WriteByte(' ')

		_, ok = expr.right.(Predicate)
		if ok {
			s.sb.WriteByte('(')
		}
		if err := s.buildExpression(expr.right); err != nil {
			return err
		}
		if ok {
			s.sb.WriteByte(')')
		}

	case Column:
		fd, ok := s.model.FieldMap[expr.name]
		if !ok {
			return fmt.Errorf("位置字段： %v", expr.name)
		}
		s.sb.WriteByte('`')
		s.sb.WriteString(fd.colName)
		s.sb.WriteByte('`')
	case Value:
		s.sb.WriteByte('?')
		if s.args == nil {
			s.args = make([]any, 0)
		}
		s.args = append(s.args, expr.val)

	default:
		return fmt.Errorf("不支持此表达式")
	}
	return nil
}
