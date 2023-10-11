package ratelimit

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

type LeakyBucketLimiter struct {
	producer *time.Ticker
}

func NewLeakyBucketLimiter(interval time.Duration) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		producer: time.NewTicker(interval),
	}
}

func (t *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		select {
		case <-ctx.Done():
			// 超时了
			err = ctx.Err()
			return
		case <-t.producer.C:
			// 漏牌桶漏进来了
			resp, err = handler(ctx, req)

		}
		return
	}
}

func (t *LeakyBucketLimiter) Close() error {
	t.producer.Stop()
	return nil
}
