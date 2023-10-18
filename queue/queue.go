package queue

import "context"

type Queue[T any] interface {
	// 入列
	Enqueue(ctx context.Context, data any) error
	// 出列
	Dequeue(ctx context.Context) (any, error)
	Len() uint64

	// 队列是否满了
	IsFull() bool
	// 队列是否为空
	IsEmpty() bool
}
