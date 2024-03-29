package orm

// Assignable 标记接口，
// 实现该接口意味着可以用于赋值语句，
// 用于在 UPDATE 和 UPSERT 中
type Assignable interface {
	assign()
}

type Assignment struct {
	column string
	val    any
}

func (a Assignment) assign() {}

func Assign(col string, val any) Assignment {
	return Assignment{
		column: col,
		val:    val,
	}
}
