package queue

import (
	"context"
	"sync"
)

type ConcurrentBlockingQueue[T any] struct {
	mutex    *sync.Mutex
	data     []T
	notFull  chan struct{}
	notEmpty chan struct{}
}

func NewConcurrentBlockingQueue[T any](maxSize int) *ConcurrentBlockingQueue[T] {
	m := &sync.Mutex{}
	return &ConcurrentBlockingQueue[T]{
		data:  make([]T, 0, maxSize),
		mutex: m,
		//notFull:
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

	}

	// 这里不用if 需要用for
	// 如果有多个G，比如G1和G2， 同时唤醒，G2直接入队，G1需要再次阻塞自己 所以使用for
	for c.IsFull() {
		// 我阻塞住我自己，知道有人唤醒我
		c.mutex.Unlock()
		select {
		case <-c.notFull:
			c.mutex.Lock()
		case <-ctx.Done():
			return ctx.Err()
		}

	}
	c.data = append(c.data, data)
	c.notEmpty <- struct{}{}
	c.mutex.Unlock()
	return nil
}

func (c *ConcurrentBlockingQueue[T]) Dequeue(ctx context.Context) (any, error) {
	c.mutex.Lock()
	for c.IsEmpty() {
		// 阻塞我自己，等待元素入队
		c.mutex.Unlock()

		select {
		case <-c.notEmpty:

		}

		//var t T
		//return t, errors.New("空的队列")
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
