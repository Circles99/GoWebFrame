package orm

import (
	"GoWebFrame/orm/errs"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestSelector_Join(t *testing.T) {
	//db := memoryDB(t)
	db, _ := NewDB()
	type Order struct {
		Id        int
		UsingCol1 string
		UsingCol2 string
	}

	type OrderDetail struct {
		OrderId int
		ItemId  int

		UsingCol1 string
		UsingCol2 string
	}

	type Item struct {
		Id int
	}

	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 虽然泛型是 Order，但是我们传入 OrderDetail
			name: "specify table",
			q:    NewSelector[Order](db).From(TableOf(&OrderDetail{})),
			wantQuery: &Query{
				SQL: "SELECT * FROM `order_detail`;",
			},
		},
		{
			name: "join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{})
				return NewSelector[Order](db).
					From(t1.Join(t2).On(t1.C("Id").Eq(t2.C("OrderId"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` AS `t1` JOIN `order_detail` ON `t1`.`id` = `order_id`);",
			},
		},
		{
			name: "multiple join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := TableOf(&Item{}).As("t3")
				return NewSelector[Order](db).
					From(t1.Join(t2).
						On(t1.C("Id").Eq(t2.C("OrderId"))).
						Join(t3).On(t2.C("ItemId").Eq(t3.C("Id"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((`order` AS `t1` JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) JOIN `item` AS `t3` ON `t2`.`item_id` = `t3`.`id`);",
			},
		},
		{
			name: "left multiple join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := TableOf(&Item{}).As("t3")
				return NewSelector[Order](db).
					From(t1.LeftJoin(t2).
						On(t1.C("Id").Eq(t2.C("OrderId"))).
						LeftJoin(t3).On(t2.C("ItemId").Eq(t3.C("Id"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((`order` AS `t1` LEFT JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) LEFT JOIN `item` AS `t3` ON `t2`.`item_id` = `t3`.`id`);",
			},
		},
		{
			name: "right multiple join",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{}).As("t2")
				t3 := TableOf(&Item{}).As("t3")
				return NewSelector[Order](db).
					From(t1.RightJoin(t2).
						On(t1.C("Id").Eq(t2.C("OrderId"))).
						RightJoin(t3).On(t2.C("ItemId").Eq(t3.C("Id"))))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM ((`order` AS `t1` RIGHT JOIN `order_detail` AS `t2` ON `t1`.`id` = `t2`.`order_id`) RIGHT JOIN `item` AS `t3` ON `t2`.`item_id` = `t3`.`id`);",
			},
		},

		{
			name: "join multiple using",
			q: func() QueryBuilder {
				t1 := TableOf(&Order{}).As("t1")
				t2 := TableOf(&OrderDetail{})
				return NewSelector[Order](db).
					From(t1.Join(t2).Using("UsingCol1", "UsingCol2"))
			}(),
			wantQuery: &Query{
				SQL: "SELECT * FROM (`order` AS `t1` JOIN `order_detail` USING (`using_col1`,`using_col2`));",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_build(t *testing.T) {
	db, _ := NewDB()
	testCase := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "no from",
			q:    NewSelector[TestModel](db),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
			wantErr: nil,
		},

		{
			name: "from",
			q:    NewSelector[TestModel](db).From(TableOf(&TestModel{})),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
			wantErr: nil,
		},
		{
			name: "where",
			q:    NewSelector[TestModel](db).From(TableOf(&TestModel{})).Where(C("FirstName").Eq("王胖子是脑残")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `first_name` = ?;",
				Args: []any{"王胖子是脑残"},
			},
			wantErr: nil,
		},

		{
			name: "where GT",
			q:    NewSelector[TestModel](db).From(TableOf(&TestModel{})).Where(C("Age").GT(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE `age` > ?;",
				Args: []any{18},
			},
			wantErr: nil,
		},

		{
			name: "where multiple GT",
			q:    NewSelector[TestModel](db).From(TableOf(&TestModel{})).Where(C("Age").GT(18), C("Id").Eq(1)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE (`age` > ?) AND (`id` = ?);",
				Args: []any{18, 1},
			},
			wantErr: nil,
		},

		{
			name: "where not",
			q:    NewSelector[TestModel](db).From(TableOf(&TestModel{})).Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` WHERE  NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},

		{
			name: "select",
			q:    NewSelector[TestModel](db).Select(C("Id")).From(TableOf(&TestModel{})).Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT `id` FROM `test_model` WHERE  NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},

		{
			name: "select alias",
			q:    NewSelector[TestModel](db).Select(C("Id").As("aliasId")).From(TableOf(&TestModel{})).Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT `id` AS `aliasId` FROM `test_model` WHERE  NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},

		{
			name: "select aggregate",
			q:    NewSelector[TestModel](db).Select(Count("Id")).From(TableOf(&TestModel{})).Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT COUNT(`id`) FROM `test_model` WHERE  NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},
		{
			name: "select aggregate as",
			q:    NewSelector[TestModel](db).Select(Count("Id").AS("aliasId")).From(TableOf(&TestModel{})).Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT COUNT(`id`) AS `aliasId` FROM `test_model` WHERE  NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},
		{
			name: "select aggregate as and where",
			q:    NewSelector[TestModel](db).Select(Count("Id").AS("aliasId")).From(TableOf(&TestModel{})).Where(Avg("Id").EQ(2)),
			wantQuery: &Query{
				SQL:  "SELECT COUNT(`id`) AS `aliasId` FROM `test_model` WHERE AVG(`id`) = ?;",
				Args: []any{2},
			},
			wantErr: nil,
		},
		{
			// 使用 RawExpr
			name: "raw expression",
			q: NewSelector[TestModel](db).Select(Raw("`id`")).
				Where(Raw("`age` < ?", 18).AsPredicate()),
			wantQuery: &Query{
				SQL:  "SELECT `id` FROM `test_model` WHERE `age` < ?;",
				Args: []any{18},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			fmt.Println(err)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_OffsetLimit(t *testing.T) {
	db, _ := NewDB()
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "offset only",
			q:    NewSelector[TestModel](db).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` OFFSET ?;",
				Args: []any{10},
			},
		},
		{
			name: "limit only",
			q:    NewSelector[TestModel](db).Limit(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` LIMIT ?;",
				Args: []any{10},
			},
		},
		{
			name: "limit offset",
			q:    NewSelector[TestModel](db).Limit(20).Offset(10),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` LIMIT ? OFFSET ?;",
				Args: []any{20, 10},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_Having(t *testing.T) {
	//db := memoryDB(t)
	db, _ := NewDB()
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 调用了，但是啥也没传
			name: "none",
			q:    NewSelector[TestModel](db).GroupBy(C("Age")).Having(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
			},
		},
		{
			// 单个条件
			name: "single",
			q: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(C("FirstName").Eq("Deng")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING `first_name` = ?;",
				Args: []any{"Deng"},
			},
		},
		{
			// 多个条件
			name: "multiple",
			q: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(C("FirstName").Eq("Deng"), C("LastName").Eq("Ming")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING (`first_name` = ?) AND (`last_name` = ?);",
				Args: []any{"Deng", "Ming"},
			},
		},
		{
			// 聚合函数
			name: "avg",
			q: NewSelector[TestModel](db).GroupBy(C("Age")).
				Having(Avg("Age").EQ(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `test_model` GROUP BY `age` HAVING AVG(`age`) = ?;",
				Args: []any{18},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			fmt.Println(query)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestSelector_GroupBy(t *testing.T) {
	db, _ := NewDB()
	testCases := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			// 调用了，但是啥也没传
			name: "none",
			q:    NewSelector[TestModel](db).GroupBy(),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model`;",
			},
		},
		{
			// 单个
			name: "single",
			q:    NewSelector[TestModel](db).GroupBy(C("Age")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`;",
			},
		},
		{
			// 多个
			name: "multiple",
			q:    NewSelector[TestModel](db).GroupBy(C("Age"), C("FirstName")),
			wantQuery: &Query{
				SQL: "SELECT * FROM `test_model` GROUP BY `age`,`first_name`;",
			},
		},
		{
			// 不存在
			name:    "invalid column",
			q:       NewSelector[TestModel](db).GroupBy(C("Invalid")),
			wantErr: errs.NewErrUnknownField("Invalid"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := tc.q.Build()
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantQuery, query)
		})
	}
}

func TestAA(T *testing.T) {
	a := &TestModel{}
	typ := reflect.TypeOf(a).Elem()

	for i := 0; i < typ.NumField(); i++ {
		fd := typ.Field(i)
		v := reflect.New(fd.Type)
		fmt.Println("V", v)
		fmt.Println("fd.Type", fd.Type)
		fmt.Printf("interface, %v", v.Kind())
		fmt.Println()
		fmt.Printf("elem, %v", v.Elem().Kind())
		fmt.Println()
	}

}

func memoryDB(t *testing.T) *DB {
	orm, err := Open("sqlite3", "file:test.db?cache=shared&mode=memory")
	if err != nil {
		t.Fatal(err)
	}
	return orm
}

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
