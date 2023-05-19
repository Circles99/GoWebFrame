package orm

import (
	"GoWebFrame/orm/errs"
	"GoWebFrame/orm/interal/model"
	"strings"
)

type builder struct {
	sb      strings.Builder
	model   *model.Model
	args    []any
	dialect Dialect
	quoter  byte
}

// buildColumn 构造列
func (b *builder) buildColumn(fd string) error {
	meta, ok := b.model.FieldMap[fd]
	if !ok {
		return errs.NewErrUnknownField(fd)
	}
	b.quote(meta.ColName)
	return nil
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.quoter)
	b.sb.WriteString(name)
	b.sb.WriteByte(b.quoter)
}

func (b *builder) addArgs(args ...any) {
	if b.args == nil {
		// 很少有查询能够超过八个参数
		// INSERT 除外
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, args...)
}

func (b *builder) raw(r RawExpr) {
	b.sb.WriteString(r.raw)
	if len(r.args) != 0 {
		b.addArgs(r.args...)
	}
}

func (b *builder) buildPredicates(ps []Predicate) error {
	// 拿出第0个, 下面循环从1开始，进行and, 二叉树
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		// 每次返回一个新的predicate,重新赋值
		p = p.And(ps[i])
	}
	return b.buildExpression(p)
}

// buildExpression
// Expression 是从 sql where 中获取的字段， model中的field 从 struct中获取
func (b *builder) buildExpression(e Expression) error {
	if e == nil {
		return nil
	}
	switch exp := e.(type) {
	case Column:
		return b.buildColumn(exp.name)
	case Aggregate:
		return b.buildAggregate(exp)
	case Value:
		b.sb.WriteByte('?')
		b.addArgs(exp.val)
	case RawExpr:
		b.sb.WriteString(exp.raw)
		if len(exp.args) != 0 {
			b.addArgs(exp.args...)
		}
	case Predicate:
		// 进行左右两边递归调用拼接值
		_, ok := exp.left.(Predicate)
		// 当只有是Predicate类型的时候才加括号
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exp.left); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}

		if exp.op == "" {
			return nil
		}

		// 拼接中间
		b.sb.WriteByte(' ')
		b.sb.WriteString(exp.op.String())
		b.sb.WriteByte(' ')

		_, ok = exp.right.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(exp.right); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}
	default:
		return errs.NewErrUnsupportedExpressionType(exp)
	}
	return nil
}

func (b *builder) buildAggregate(a Aggregate) error {
	b.sb.WriteString(a.fn)
	b.sb.WriteByte('(')
	err := b.buildColumn(a.arg)
	if err != nil {
		return err
	}
	b.sb.WriteByte(')')
	if a.alias != "" {
		b.buildAs(a.alias)
	}
	return nil
}

func (b *builder) buildAs(alias string) {
	if alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(alias)
	}
}
