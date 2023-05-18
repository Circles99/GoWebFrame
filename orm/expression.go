package orm

type RawExpr struct {
	raw  string
	args []interface{}
}

func (r RawExpr) selectable() {}
func (r RawExpr) expr()       {}

// AsPredicate 转为Predicate, 用于where中的表达式
func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

func Raw(expr string, args ...interface{}) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}
