//go:build: e2e
package ratelimit

import (
	"GoWebFrame/micro/example/gen"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"testing"
	"time"
)

func TestRedisFixWindow_BuildServerInterceptor(t *testing.T) {

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	interceptor := NewRedisFixWindowLimiter(rdb, "user-service", time.Second*3, 1).BuildServerInterceptor()

	cnt := 0
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		cnt++
		return &gen.GetByIdResp{}, nil
	}

	resp, err := interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdResp{}, resp)

	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.Equal(t, errors.New("触发瓶颈了"), err)
	require.Nil(t, resp)

	// 睡3秒，确保窗口新建了
	time.Sleep(3 * time.Second)
	resp, err = interceptor(context.Background(), &gen.GetByIdReq{}, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &gen.GetByIdResp{}, resp)
}
