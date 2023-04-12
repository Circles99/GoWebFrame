package orm

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

type Selector[T any] struct {
	sb    *strings.Builder // sb在指针中也会引起复制，所以需要获取指针
	tbl   string
	where []Predicate
	args  []any
}

func NewSelector[T any]() *Selector[T] {
	return &Selector[T]{}
}

// From 加入表名，为了链式调用返回Selector[T]
func (s *Selector[T]) From(tbl string) *Selector[T] {
	s.tbl = tbl
	return s
}

func (s *Selector[T]) Where() *Selector[T] {
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	s.sb = &strings.Builder{}

	s.sb.WriteString("SELECT * FROM ")

	if s.tbl == "" {
		var t T
		// 获取类型
		typ := reflect.TypeOf(t)
		// 获取名字
		s.sb.WriteString("`")
		s.sb.WriteString(typ.Name())
		s.sb.WriteString("`")
	} else {
		s.sb.WriteString(s.tbl)
	}

	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE")
		// 拿出第0个, 下面循环从1开始，进行and, 二叉树
		p := s.where[0]
		for i := 1; i < len(s.where); i++ {
			p.And(s.where[1])
		}
		err := s.buildExpression(p)
		if err != nil {
			return nil, err
		}
	}

	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: nil,
	}, nil
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

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
		s.sb.WriteByte('`')
		s.sb.WriteString(expr.name)
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
