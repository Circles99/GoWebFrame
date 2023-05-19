package orm

import (
	"GoWebFrame/orm/errs"
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

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
			q:    NewSelector[TestModel](db).From("`TestModel`"),
			wantQuery: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
			wantErr: nil,
		},
		{
			name: "where",
			q:    NewSelector[TestModel](db).From("`TestModel`").Where(C("FirstName").Eq("王胖子是脑残")),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE `first_name` = ?;",
				Args: []any{"王胖子是脑残"},
			},
			wantErr: nil,
		},

		{
			name: "where GT",
			q:    NewSelector[TestModel](db).From("`TestModel`").Where(C("Age").GT(18)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE `age` > ?;",
				Args: []any{18},
			},
			wantErr: nil,
		},

		{
			name: "where multiple GT",
			q:    NewSelector[TestModel](db).From("`TestModel`").Where(C("Age").GT(18), C("Id").Eq(1)),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE (`age` > ?) AND (`id` = ?);",
				Args: []any{18, 1},
			},
			wantErr: nil,
		},

		{
			name: "where not",
			q:    NewSelector[TestModel](db).From("`TestModel`").Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT * FROM `TestModel` WHERE NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},

		{
			name: "select",
			q:    NewSelector[TestModel](db).Select(C("Id")).From("`TestModel`").Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT `id` FROM `TestModel` WHERE  NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},

		{
			name: "select alias",
			q:    NewSelector[TestModel](db).Select(C("Id").As("aliasId")).From("`TestModel`").Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT `id` AS `aliasId` FROM `TestModel` WHERE  NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},

		{
			name: "select aggregate",
			q:    NewSelector[TestModel](db).Select(Count("Id")).From("`TestModel`").Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT COUNT(`id`) FROM `TestModel` WHERE  NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},
		{
			name: "select aggregate as",
			q:    NewSelector[TestModel](db).Select(Count("Id").AS("aliasId")).From("`TestModel`").Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "SELECT COUNT(`id`) AS `aliasId` FROM `TestModel` WHERE  NOT (`id` = ?);",
				Args: []any{2},
			},
			wantErr: nil,
		},
		{
			name: "select aggregate as and where",
			q:    NewSelector[TestModel](db).Select(Count("Id").AS("aliasId")).From("`TestModel`").Where(Avg("Id").EQ(2)),
			wantQuery: &Query{
				SQL:  "SELECT COUNT(`id`) AS `aliasId` FROM `TestModel` WHERE AVG(`id`) = ?;",
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
	db := memoryDB(t)
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
