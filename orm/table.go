package orm

type TableReference interface {
	tableAlias() string
}

// Table 普通表
type Table struct {
	entity any
	alias  string
}

func TableOf(entity any) Table {
	return Table{
		entity: entity,
	}
}

func (t Table) C(name string) Column {
	return Column{
		name:  name,
		table: t,
	}
}

func (t Table) As(alias string) Table {
	t.alias = alias
	return t
}

func (t Table) tableAlias() string {
	return t.alias
}

func (t Table) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "JOIN",
	}
}

func (t Table) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (t Table) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  t,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

// Join 连接表, table利用join方法进来之后可在继续join,所以join结构体也需要这三个join方法
type Join struct {
	left  TableReference
	right TableReference
	typ   string
	on    []Predicate
	using []string
}

func (j Join) Join(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "JOIN",
	}
}

func (j Join) LeftJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "LEFT JOIN",
	}
}

func (j Join) RightJoin(target TableReference) *JoinBuilder {
	return &JoinBuilder{
		left:  j,
		right: target,
		typ:   "RIGHT JOIN",
	}
}

func (j Join) tableAlias() string {
	return ""
}

// JoinBuilder joinbuilder和onConflictBuilder的设计理念一样，这里为了他table JOIN之后只能用ON和using,所以引入joinbuilder
type JoinBuilder struct {
	left  TableReference
	right TableReference
	typ   string
}

func (j *JoinBuilder) On(ps ...Predicate) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		on:    ps,
	}
}

func (j *JoinBuilder) Using(cols ...string) Join {
	return Join{
		left:  j.left,
		right: j.right,
		typ:   j.typ,
		using: cols,
	}
}
