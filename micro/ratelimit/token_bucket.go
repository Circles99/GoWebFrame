package ratelimit

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"time"
)

type TokenBucketLimiter struct {
	tokens chan struct{}
	close  chan struct{}
}

func NewTokenBucketLimiter(cap int, interval time.Duration) *TokenBucketLimiter {
	// interval 是隔多久产生一个令牌
	ch := make(chan struct{}, cap)
	closeCh := make(chan struct{})

	producer := time.NewTicker(interval)
	go func() {
		defer producer.Stop()
		for {
			select {
			case <-producer.C:
				// 这个地方使用select是为了如果一直没人取走这个令牌，但有人调用了close，因为阻塞住了，没人能够取到这个令牌了
				select {
				case ch <- struct{}{}:
				default:
					//没人取令牌
				}
			case <-closeCh:
				return
			}
		}
	}()

	return &TokenBucketLimiter{
		// cap是容量，在这个令牌桶里 最多攒下几个令牌
		tokens: ch,
		close:  closeCh,
	}
}

func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		select {
		case <-t.close:
			// 已经关闭故障检测了
			//resp, err = handler(ctx, req)
			err = errors.New("缺乏保护，拒绝请求")
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-t.tokens:
			// 要在这拿到令牌
			resp, err = handler(ctx, req)
		default:
			err = errors.New("到达瓶颈")
		}

		return
	}
}

func (t *TokenBucketLimiter) Close() error {
	close(t.close)
	return nil
}

//func (t *TokenBucketLimiter) BuildClientInterceptor() grpc.UnaryClientInterceptor {
//
//}
