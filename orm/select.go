package orm

import (
	"GoWebFrame/orm/interal/model"
	"context"
	"fmt"
	"strings"
)

type Selector[T any] struct {
	sb      *strings.Builder // sb在指针中也会引起复制，所以需要获取指针
	tbl     string
	model   *model.Model
	where   []Predicate
	columns []selectedAlias
	args    []any
	db      *DB
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

// Select 加入需要查询的字段
func (s *Selector[T]) Select(cols ...selectedAlias) *Selector[T] {
	s.columns = cols
	return s
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

	s.sb.WriteString("SELECT ")

	// 进行拼接
	if len(s.columns) > 0 {
		for i, c := range s.columns {
			if i > 0 {
				s.sb.WriteByte(',')
			}
			switch val := c.(type) {
			case Column:
				err = s.buildColumn(val)
				if err != nil {
					return nil, err
				}
			case Aggregate:
				if err = s.buildAggregate(val); err != nil {
					return nil, err
				}
			case RawExpr:
				s.sb.WriteString(val.raw)
				if len(val.args) > 0 {
					s.args = append(s.args, val.args...)
				}
			}
		}
	} else {
		s.sb.WriteByte('*')
	}

	s.sb.WriteString(" FROM ")

	if s.tbl == "" {
		//var t T
		//// 获取类型
		//typ := reflect.TypeOf(t)
		// 获取名字
		s.sb.WriteString("`")
		//s.sb.WriteString(typ.Name ())
		s.sb.WriteString(s.model.TableName)
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

	t := new(T)
	val := s.db.ValCreator(t, s.model)
	// 在这里灵活切换反射或者 unsafe

	return t, val.SetColumns(rows)
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

		if expr.op == "" {
			return nil
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
		s.sb.WriteString(fd.ColName)
		s.sb.WriteByte('`')
	case Value:
		s.sb.WriteByte('?')
		if s.args == nil {
			s.args = make([]any, 0)
		}
		s.args = append(s.args, expr.val)
	case Aggregate:
		if err := s.buildAggregate(expr); err != nil {
			return err
		}
	case RawExpr:
		s.sb.WriteString(expr.raw)
		if len(expr.args) > 0 {
			s.args = append(s.args, expr.args...)
		}

	default:
		return fmt.Errorf("不支持此表达式")
	}
	return nil
}

func (s *Selector[T]) buildColumn(c Column) error {
	fd, ok := s.model.FieldMap[c.name]
	if !ok {
		return fmt.Errorf("位置字段： %v", c.name)
	}
	s.sb.WriteByte('`')
	s.sb.WriteString(fd.ColName)
	s.sb.WriteByte('`')

	if c.alias != "" {
		s.buildAlias(c.alias)
	}

	return nil
}

func (s *Selector[T]) buildAlias(alias string) {
	s.sb.WriteString(" AS ")
	s.sb.WriteByte('`')
	s.sb.WriteString(alias)
	s.sb.WriteByte('`')
}

func (s *Selector[T]) buildAggregate(a Aggregate) error {
	s.sb.WriteString(a.fn)
	s.sb.WriteByte('(')
	err := s.buildColumn(Column{
		name: a.arg,
	})
	if err != nil {
		return err
	}
	s.sb.WriteByte(')')
	if a.alias != "" {
		s.buildAlias(a.alias)
	}
	return nil
}
