package queue

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

func TestNewConcurrentBlockingQueue(t *testing.T) {
	q := NewConcurrentBlockingQueue[int](1000)

	for i := 0; i < 20; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			q.Enqueue(ctx, rand.Int())
			cancel()
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			q.Dequeue(ctx)
			cancel()
		}()
	}
}
