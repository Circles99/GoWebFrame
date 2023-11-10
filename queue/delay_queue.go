package queue

import (
	"context"
	"sync"
	"time"
)

type DelayQueue[T Delayable] struct {
	q     *PriorityQueue[T]
	mutex sync.Mutex
}

func NewDelayQueue[T Delayable](capacity int) *DelayQueue[T] {
	return &DelayQueue[T]{
		q: NewPriorityQueue[T](10, func(src T, dst T) int {
			srcDelay := src.Delay()
			dstDelay := dst.Delay()
			if srcDelay < dstDelay {
				return -1
			} else if srcDelay == dstDelay {
				return 0
			}
			return 1
		}),
	}
}

// 入队和并发阻塞队列没太大区别
func (c *DelayQueue[T]) Enqueue(ctx context.Context, data T) error {

}

// 出队
// 1：delay返回小于等于0的时候出队
// 2: 队首的delay>0 要是sleep，等待delay()降下去
// 3.如果正在sleep的过程，有新元素来了，并且delay() = 200 比你正在sleep的时间还要短
// 你要调整你的sleep时间
// 4： 如果sleep的时间还没到，就超时了，那么就返回
// sleep的本质就是阻塞，可以用time.sleep， 也可以用channel
func (c *DelayQueue[T]) Dequeue(ctx context.Context) (T, error) {

}

type Delayable interface {
	Delay() time.Duration
}
