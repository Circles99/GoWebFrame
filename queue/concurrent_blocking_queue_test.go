package queue

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestNewConcurrentBlockingQueue(t *testing.T) {
	// 只能确保不死锁
	q := NewConcurrentBlockingQueue[int](1000)

	var wg sync.WaitGroup
	wg.Add(30)
	for i := 0; i < 20; i++ {
		go func() {
			for {
				for j := 0; j < 1000; j++ {
					// 没有办法校验这里面的中间结果
					ctx, cancel := context.WithTimeout(context.Background(), time.Second)
					_ = q.Enqueue(ctx, rand.Int())
					cancel()
				}
			}
			wg.Done()

		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 1000; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				_, _ = q.Dequeue(ctx)
				cancel()
			}

			wg.Done()
		}()
	}

	wg.Wait()
}
