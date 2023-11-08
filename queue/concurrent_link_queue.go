package queue

import "context"

// 无锁队列的核心：CAS + 自旋
// cas:compare and swap 比较并交换
type ConcurrentLinkedQueue[T any] struct {
	head *node
	tail *node
}

func (c *ConcurrentLinkedQueue[T]) Enqueue(ctx context.Context, data T) error {
	// 入队
	c.tail.next = &node{
		val: data,
	}

	c.tail = c.tail.next
	return nil
}

func (c *ConcurrentLinkedQueue[T]) Dequeue(ctx context.Context) (T, error) {
	// 出队
	head := c.head
	c.head = c.head.next

	return head.val.(T), nil
}

func (c *ConcurrentLinkedQueue[T]) Len() uint64 {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkedQueue[T]) IsFull() bool {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentLinkedQueue[T]) IsEmpty() bool {
	//TODO implement me
	panic("implement me")
}

type node struct {
	next *node
	val  any
}
