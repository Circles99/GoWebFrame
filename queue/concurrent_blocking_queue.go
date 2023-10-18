package queue

import (
	"context"
	"errors"
	"sync"
)

type ConcurrentBlockingQueue[T any] struct {
	mutex   *sync.Mutex
	data    []T
	notFull *sync.Cond
}

func NewConcurrentBlockingQueue[T any](maxSize int) *ConcurrentBlockingQueue[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueue[T]{
		data:    make([]T, 0, maxSize),
		mutex:   m,
		notFull: sync.NewCond(m),
	}
}

func (c *ConcurrentBlockingQueue[T]) Enqueue(ctx context.Context, data any) error {
	c.mutex.Lock()
	// 全是bug
	// 这里需要自己释放锁，如果不是放，另外一边是拿不到
	select {
	case <-ctx.Done():
		c.mutex.Unlock()
		return ctx.Err()
	default:
		if c.IsFull() {
			// 我阻塞住我自己，知道有人唤醒我
			c.notFull.Wait()
		}
	}

	c.data = append(c.data, data)
	c.mutex.Unlock()
	return nil
}

func (c *ConcurrentBlockingQueue[T]) Dequeue(ctx context.Context) (any, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.IsEmpty() {
		var t T
		return t, errors.New("空的队列")
	}

	// 队首
	t := c.data[0]
	c.data = c.data[1:]
	return t, nil

}

func (c *ConcurrentBlockingQueue[T]) Len() uint64 {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentBlockingQueue[T]) IsFull() bool {
	//TODO implement me
	panic("implement me")
}

func (c *ConcurrentBlockingQueue[T]) IsEmpty() bool {
	//TODO implement me
	panic("implement me")
}
