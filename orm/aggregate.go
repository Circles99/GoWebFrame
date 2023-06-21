package orm

// Package orm
// @Description: sql函数接口

type Aggregate struct {
	table TableReference
	fn    string // 函数
	arg   string // 参数
	alias string // 别名
}

func (a Aggregate) expr() {}

func (a Aggregate) selectable() {}

func (a Aggregate) GT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opGT,
		right: exprOf(arg),
	}
}

func (a Aggregate) LT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opLT,
		right: exprOf(arg),
	}
}

func (a Aggregate) EQ(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opEQ,
		right: exprOf(arg),
	}
}

func (a Aggregate) AS(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

func Avg(a string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: a,
	}
}

func Max(a string) Aggregate {
	return Aggregate{
		fn:  "MAX",
		arg: a,
	}
}

func Min(a string) Aggregate {
	return Aggregate{
		fn:  "MIN",
		arg: a,
	}
}

func Count(a string) Aggregate {
	return Aggregate{
		fn:  "COUNT",
		arg: a,
	}
}

func Sum(c string) Aggregate {
	return Aggregate{
		fn:  "SUM",
		arg: c,
	}
}
