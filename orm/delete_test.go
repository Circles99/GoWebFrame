package orm

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDelete_build(t *testing.T) {
	db, _ := NewDB()
	testCase := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "no from",
			q:    NewDelete[TestModel](db),
			wantQuery: &Query{
				SQL: "DELETE FROM `test_model`;",
			},
			wantErr: nil,
		},

		{
			name: "from",
			q:    NewDelete[TestModel](db).From("`TestModel1`"),
			wantQuery: &Query{
				SQL: "DELETE FROM `TestModel1`;",
			},
			wantErr: nil,
		},
		{
			name: "where",
			q:    NewDelete[TestModel](db).From("`TestModel`").Where(C("FirstName").Eq("王胖子是脑残")),
			wantQuery: &Query{
				SQL:  "DELETE FROM `TestModel` WHERE `first_name` = ?;",
				Args: []any{"王胖子是脑残"},
			},
			wantErr: nil,
		},

		{
			name: "where GT",
			q:    NewDelete[TestModel](db).From("`TestModel`").Where(C("Age").GT(18)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `TestModel` WHERE `age` > ?;",
				Args: []any{18},
			},
			wantErr: nil,
		},

		{
			name: "where multiple GT",
			q:    NewDelete[TestModel](db).From("`TestModel`").Where(C("Age").GT(18), C("Id").Eq(1)),
			wantQuery: &Query{
				SQL:  "DELETE FROM `TestModel` WHERE (`age` > ?) AND (`id` = ?);",
				Args: []any{18, 1},
			},
			wantErr: nil,
		},

		{
			name: "where not",
			q:    NewDelete[TestModel](db).From("`TestModel`").Where(Not(C("Id").Eq(2))),
			wantQuery: &Query{
				SQL:  "DELETE FROM `TestModel` WHERE  NOT (`id` = ?);",
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
