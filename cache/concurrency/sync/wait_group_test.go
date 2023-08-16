package main

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestWaitGroup(t *testing.T) {
	wg := &sync.WaitGroup{}
	var result int64

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			atomic.AddInt64(&result, int64(i))
		}(i)
	}
	wg.Wait()
	t.Log(result)
}
