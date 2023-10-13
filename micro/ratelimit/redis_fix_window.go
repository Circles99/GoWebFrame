package ratelimit

import (
	"context"
	_ "embed"
	"errors"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"time"
)

// go:embed lua/fix_window.lua
var luaFixWindow string

type RedisFixWindowLimiter struct {
	client   redis.Cmdable
	service  string
	interval time.Duration
	// 阈值
	rate int
}

func NewRedisFixWindowLimiter(client redis.Cmdable, service string, interval time.Duration, rate int) *RedisFixWindowLimiter {
	return &RedisFixWindowLimiter{
		client:   client,
		service:  service,
		interval: interval,
		rate:     rate,
	}
}

func (t *RedisFixWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 我预期lua脚本返回一个bool值， 告诉我是否限流
		// 使用fullMethod 就是单一方法上的限流，比如getById
		// 使用服务名，就是在单一服务上 users.UserService
		limit, err := t.limit(ctx, info.FullMethod)

		if err != nil {
			return
		}

		if limit {
			err = errors.New("触及了限流")
			return
		}

		resp, err = handler(ctx, req)
		return
	}
}

func (t *RedisFixWindowLimiter) limit(ctx context.Context, service string) (bool, error) {
	// 也可以使用t.service
	return t.client.Eval(ctx, luaFixWindow, []string{service}, t.interval.Milliseconds(), t.rate).Bool()
}
