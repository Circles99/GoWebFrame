package orm

import (
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
