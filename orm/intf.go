package orm

import "context"

// Executor 认为所有写操作都是一个exec执行，谁需要写就去实现改interface
type Executor interface {
	Exec(ctx context.Context) error
}

// Querier 同上，select就是一个查询操作，谁需要读就实现该interface
type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) ([]*T, error)
}

// Query 用来build sql数据
type Query struct {
	SQL  string
	Args []any
}

type QueryBuilder interface {
	Build() (*Query, error)
}
