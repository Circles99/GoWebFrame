package orm

type op string

const (
	opEQ = "="
	opLT = "<"
	opGT = ">"

	opNOT = "NOT"
	opAND = "AND"
	opOR  = "OR"
)

func (o op) String() string {
	return string(o)
}

type Expression interface {
	expr()
}

// 接受结构化的Predicate进行输入

type Predicate struct {
	left  Expression // 左边匹配的column
	op    op         // 标记符号
	right Expression // 右边值
}

// 标记接口
func (p Predicate) expr() {}

func exprOf(e any) Expression {
	//value 比较特殊, 直接是值形式，其他的都已表达式返回
	switch val := e.(type) {
	case Expression:
		return val
	default:
		return Value{val: val}
	}
}

// Not(C("id").Eq(12))
// NOT (id = ?), 12
func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNOT,
		right: p,
	}
}

// C("id").Eq(12).And(C("name").Eq("Tom"))
func (p1 Predicate) And(p2 Predicate) Predicate {
	return Predicate{
		left:  p1,
		op:    opAND,
		right: p2,
	}
}

// C("id").Eq(12).Or(C("name").Eq("Tom"))
func (p1 Predicate) Or(p2 Predicate) Predicate {
	return Predicate{
		left:  p1,
		op:    opOR,
		right: p2,
	}
}
