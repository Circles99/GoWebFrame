package orm

type Column struct {
	name    string
	alias   string
	orderBy string
	table   TableReference
}

// 标记接口
func (Column) expr() {}

// 标记接口
func (Column) selectable() {}

func (Column) assign() {}

// As 设置别名
func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
	}
}

// Desc 排序倒叙
func (c Column) Desc(orderBy string) Column {
	return Column{
		name:    c.name,
		orderBy: orderBy,
	}
}

type Value struct {
	val any
}

func (Value) expr() {}

func C(name string) Column {
	return Column{name: name}
}

func (c Column) Eq(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opEQ,
		right: exprOf(val),
	}
}

func (c Column) GT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: exprOf(val),
	}
}

func (c Column) LT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: exprOf(val),
	}
}
