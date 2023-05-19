package orm

type Column struct {
	name    string
	alias   string
	orderBy string
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
