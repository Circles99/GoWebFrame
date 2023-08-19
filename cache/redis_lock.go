package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
	"time"
)

var (
	//go:embed lua/unlock.lua
	luaUnlock string
	//go:embed lua/refresh.lua
	luaRefresh string
	//go:embed lua/lock.lua
	luaLock                  string
	ErrorFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
	ErrLockNotHold           = errors.New("rlock: 未持有锁")
)

// Client 就是对redis.Cmdable的二次封装
type Client struct {
	client redis.Cmdable
	g      singleflight.Group
}

func (c *Client) SingleflightLock(ctx context.Context, key string, expiration time.Duration, timeout time.Duration, retry RetryStrategy) (*Lock, error) {
	for {
		flag := false
		resChan := c.g.DoChan(key, func() (interface{}, error) {
			flag = true
			return c.Lock(ctx, key, expiration, timeout, retry)
		})
		select {
		case res := <-resChan:
			if flag {
				// 确保下一个循环 另一个goroutine触发
				c.g.Forget(key)
				if res.Err != nil {
					return nil, res.Err
				}
				return res.Val.(*Lock), nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) Lock(ctx context.Context, key string, expiration time.Duration, timeout time.Duration, retry RetryStrategy) (*Lock, error) {

	var timer *time.Timer

	val := uuid.New().String()

	for {
		lctx, cancel := context.WithTimeout(ctx, timeout)
		res, err := c.client.Eval(lctx, luaLock, []string{key}, val, expiration.Seconds()).Result()
		cancel()
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}

		if res == "OK" {
			//加锁成功了
			return &Lock{
				client:     c.client,
				key:        key,
				value:      val,
				expiration: expiration,
				UnlockChan: make(chan struct{}, 1),
			}, nil
		}

		interval, ok := retry.Next()
		if !ok {
			return nil, fmt.Errorf("redis-lock 超出重试限制 %s", ErrorFailedToPreemptLock)
		}
		if timer == nil {
			timer = time.NewTimer(interval)
		} else {
			timer.Reset(interval)
		}

		select {
		case <-timer.C:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (c *Client) TryLock(ctx context.Context, key string, expiration time.Duration) (*Lock, error) {
	// 利用uuid去判断解锁时，锁是不是自己设置的锁
	val := uuid.New().String()

	ok, err := c.client.SetNX(ctx, key, val, expiration).Result()
	if err != nil {
		return nil, err
	}

	if !ok {
		// 别人抢到了锁
		return nil, ErrorFailedToPreemptLock
	}

	return &Lock{
		client:     c.client,
		key:        key,
		value:      val,
		expiration: expiration,
		UnlockChan: make(chan struct{}, 1),
	}, nil
}

type Lock struct {
	client     redis.Cmdable
	key        string
	value      string
	expiration time.Duration
	UnlockChan chan struct{}
}

// AutoRefresh 自动续约， 总体还是需要用户自己管
func (l *Lock) AutoRefresh(interval time.Duration, timeout time.Duration) error {
	timeoutChan := make(chan struct{}, 1)
	// 间隔多久续约一次
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			// 出现了error咋 办
			err := l.Refresh(ctx)
			cancel()

			if err == context.DeadlineExceeded {
				timeoutChan <- struct{}{}
				continue
			}

			if err != nil {
				return err
			}
		case <-timeoutChan:
			// 超时了重写调动刷新
			// 可以用次数控制住刷新
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			err := l.Refresh(ctx)
			cancel()

			if err == context.DeadlineExceeded {
				timeoutChan <- struct{}{}
				continue
			}

			if err != nil {
				return err
			}

		case <-l.UnlockChan:
			return nil
		}
	}
}

func (l *Lock) Refresh(ctx context.Context) error {

	res, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.value, l.expiration.Seconds()).Int64()
	if err != nil {
		return err
	}

	if res != 1 {
		return ErrLockNotHold
	}

	return nil
}

func (l *Lock) UnLock(ctx context.Context) error {

	// 使用lua脚本，必须保证其步骤在同一个原子操作中完成
	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.value).Int64()

	defer func() {
		// 可使用once保护一下这个只允许呗调用一次
		//close(l.UnlockChan)
		select {
		case l.UnlockChan <- struct{}{}:
		default:
			// 说明没有人调用
		}

	}()

	if err != nil {
		return err
	}

	if res != 1 {
		return ErrLockNotHold
	}

	//// 把键值对删掉
	//cnt, err := l.client.Del(ctx, l.key).Result()
	//if err != nil {
	//	return err
	//}
	//
	//if cnt != 1 {
	//	// 代表你家的锁过期了
	//	return errors.New("redis-lock， 解锁失败")
	//}
	return nil
}
