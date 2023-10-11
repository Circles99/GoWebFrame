package ratelimit

import (
	"container/list"
	"context"
	"errors"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type SlideWindowLimiter struct {
	queue    *list.List
	interval int64 // 间隔
	rate     int
	mutex    sync.Mutex
}

func NewSlideWindowLimiter(interval time.Duration, rate int) *SlideWindowLimiter {
	return &SlideWindowLimiter{
		interval: interval.Nanoseconds(),
		queue:    list.New(),
		rate:     rate,
	}
}

func (t *SlideWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 要知道我当前这个窗口，处理了几个请求
		now := time.Now().UnixNano()

		// 假如当前10点17， 减去1分钟， 10点16之前的数据全部删除掉， boundary就是10点16
		boundary := now - t.interval

		// 快路径，窗口没满直接调用就行
		t.mutex.Lock()
		length := t.queue.Len()
		if length < t.rate {
			resp, err = handler(ctx, req)
			t.queue.PushBack(now)
			t.mutex.Unlock()
			return
		}

		// 慢路径
		// 从队首取出来
		timestamp := t.queue.Front()

		// 把线之前的请求都删除掉
		// 这个循环把所有不在窗口内的数据都删除掉了
		for timestamp != nil && timestamp.Value.(int64) < boundary {
			t.queue.Remove(timestamp)
			timestamp = t.queue.Front()
		}

		length = t.queue.Len()
		t.mutex.Unlock()
		// 到达临街点
		if length >= t.rate {
			err = errors.New("到达瓶颈")
			return
		}

		resp, err = handler(ctx, req)
		// 记住当前请求的时间戳
		t.queue.PushBack(now)
		return
	}
}
