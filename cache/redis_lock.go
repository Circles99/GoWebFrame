package cache

import (
	"context"
	_ "embed"
	"errors"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	//go:embed lua/unlock.lua
	luaUnlock string
	//go:embed lua/refresh.lua
	luaRefresh               string
	ErrorFailedToPreemptLock = errors.New("redis-lock: 抢锁失败")
	ErrLockNotHold           = errors.New("rlock: 未持有锁")
)

// Client 就是对redis.Cmdable的二次封装
type Client struct {
	client redis.Cmdable
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
	}, nil
}

type Lock struct {
	client     redis.Cmdable
	key        string
	value      string
	expiration time.Duration
}

func (l Lock) Refresh(ctx context.Context) error {

	res, err := l.client.Eval(ctx, luaRefresh, []string{l.key}, l.value, l.expiration.Seconds()).Int64()
	if err != nil {
		return err
	}

	if res != 1 {
		return ErrLockNotHold
	}

	return nil
}

func (l Lock) UnLock(ctx context.Context) error {

	// 使用lua脚本，必须保证其步骤在同一个原子操作中完成

	res, err := l.client.Eval(ctx, luaUnlock, []string{l.key}, l.value).Int64()
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
