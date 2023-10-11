package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync/atomic"
	"time"
)

type FixWindowLimiter struct {
	// 窗口起始时间
	timestamp int64
	// 窗口大小
	interval int64

	// 在这个窗口内 允许通过的最大数量
	rate int64

	cnt int64

	//mutex sync.Mutex
}

func NewFixWindowLimiter(interval time.Duration, rate int64) *FixWindowLimiter {
	return &FixWindowLimiter{
		timestamp: time.Now().UnixNano(),
		interval:  interval.Nanoseconds(),
		rate:      rate,
	}
}

func (t *FixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		// 考虑t.cnt重置的问题
		current := time.Now().UnixNano()
		timestamp := atomic.LoadInt64(&t.timestamp)
		cnt := atomic.LoadInt64(&t.cnt)
		if timestamp+t.interval < current {
			// 意味着是一个新窗口
			//t.timestamp = current
			if atomic.CompareAndSwapInt64(&t.timestamp, timestamp, current) {
				//atomic.StoreInt64(&t.cnt, 0)
				atomic.CompareAndSwapInt64(&t.cnt, cnt, 0)
			}
		}

		cnt = atomic.AddInt64(&t.cnt, 1)

		if cnt >= t.rate {
			err = errors.New("触发瓶颈了")
			return
		}

		resp, err = handler(ctx, req)
		return
	}
}

//func (t *FixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
//	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
//
//		t.mutex.Lock()
//		// 考虑t.cnt重置的问题
//		current := time.Now().UnixNano()
//		if t.timestamp+t.interval < current {
//			// 意味着是一个新窗口
//			t.timestamp = current
//			t.cnt = 0
//		}
//
//		if t.cnt >= t.rate {
//			err = errors.New("触发瓶颈了")
//
//			t.mutex.Unlock()
//			return
//		}
//
//		t.cnt++
//		t.mutex.Unlock()
//		resp, err = handler(ctx, req)
//		return
//	}
//}
