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

// 接受结构化的Predicate进行输入

type Predicate struct {
	left  Expression // 左边匹配的column
	op    op         // 标记符号
	right Expression // 右边值
}

// 标记接口
func (p Predicate) expr() {}

type Column struct {
	name string
}

// 标记接口
func (Column) expr() {}

type Value struct {
	val any
}

func (Value) expr() {}

func C(name string) Column {
	return Column{name: name}
}

func (c Column) Eq(value any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: Value{val: value},
	}
}

func (c Column) GT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: Value{val: val},
	}
}

func (c Column) LT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: Value{val: val},
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
