package orm

import (
	model2 "GoWebFrame/orm/interal/model"
	"context"
	"errors"
	"fmt"
	"strings"
)

type Deleter[T any] struct {
	sb        *strings.Builder
	tableName string
	db        *DB
	model     *model2.Model
	where     []Predicate
	agrs      []any
}

func NewDelete[T any](db *DB) *Deleter[T] {
	return &Deleter[T]{
		db: db,
	}
}

func (d *Deleter[T]) From(tableName string) *Deleter[T] {
	d.tableName = tableName
	return d
}

func (d *Deleter[T]) Where(p ...Predicate) *Deleter[T] {
	d.where = append(d.where, p...)
	return d
}

func (d *Deleter[T]) Build() (*Query, error) {
	var (
		t   T
		err error
	)
	d.sb = &strings.Builder{}
	d.model, err = d.db.r.Get(&t)
	if err != nil {
		return nil, err
	}

	d.sb.WriteString("DELETE FROM ")

	if d.tableName == "" {
		d.sb.WriteString("`")
		d.sb.WriteString(d.model.TableName)
		d.sb.WriteString("`")
	} else {
		d.sb.WriteString(d.tableName)
	}

	if len(d.where) > 0 {
		d.sb.WriteString(" WHERE ")
		// 进行Predicate的拼接。拿出第一个来
		p := d.where[0]
		// 进行循环开始拼接
		for i := 1; i < len(d.where); i++ {
			// 用and进行连接， p = left  d.where[i] = right
			p = p.And(d.where[i])
		}
		err = d.buildExpression(p)
		if err != nil {
			return nil, err
		}
	}
	d.sb.WriteString(";")
	return &Query{
		SQL:  d.sb.String(),
		Args: d.agrs,
	}, nil

}

// buildExpression 构造表达式
// Expression 这里传入Expression接口 会因为 上述Predicate， value， Column 都是 Expression的实现
func (d *Deleter[T]) buildExpression(e Expression) error {
	switch expr := e.(type) {
	case nil:
		// Predicate 为nil的情况下
		return nil
	case Predicate:
		_, ok := expr.left.(Predicate)
		// 不是Predicate 不会加括号
		if ok {
			d.sb.WriteByte('(')
		}
		// 对左边表达式进行数据读取
		if err := d.buildExpression(expr.left); err != nil {
			return err
		}
		if ok {
			d.sb.WriteByte(')')
		}

		// 加入中间符号
		d.sb.WriteByte(' ')
		d.sb.WriteString(expr.op.String())
		d.sb.WriteByte(' ')

		_, ok = expr.right.(Predicate)
		if ok {
			d.sb.WriteByte('(')
		}
		// 对右边表达式进行数据读取
		if err := d.buildExpression(expr.right); err != nil {
			return err
		}

		if ok {
			d.sb.WriteByte(')')
		}
	case Column:
		m, ok := d.model.FieldMap[expr.name]
		if !ok {
			return errors.New("找不到此字段")
		}

		d.sb.WriteByte('`')
		d.sb.WriteString(m.ColName)
		d.sb.WriteByte('`')

	case Value:
		d.sb.WriteString("?")
		if d.agrs == nil {
			d.agrs = make([]any, 0)
		}

		d.agrs = append(d.agrs, expr.val)

	default:
		return fmt.Errorf("orm:不支持此表达式")
	}
	return nil

}

func (d *Deleter[T]) Exec(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
