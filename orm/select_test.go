package orm

import (
	"database/sql"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelector_build(t *testing.T) {
	testCase := []struct {
		name      string
		q         QueryBuilder
		wantQuery *Query
		wantErr   error
	}{
		{
			name: "no from",
			q:    &Selector[TestModel]{},
			wantQuery: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
			wantErr: nil,
		},

		{
			name: "from",
			q:    (&Selector[TestModel]{}).From("`TestModel`"),
			wantQuery: &Query{
				SQL: "SELECT * FROM `TestModel`;",
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCase {
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

type TestModel struct {
	Id        int64
	FirstName string
	Age       int8
	LastName  *sql.NullString
}
