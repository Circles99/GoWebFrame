package concurrency

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestContext(t *testing.T) {
	// 一般是链路的起点
	ctx := context.Background()
	// 在你不确定context啥时候用的时候，用todo
	//context.TODO()
	ctx = context.WithValue(ctx, "mykey1", "111")

	ctx, cancel := context.WithCancel(ctx)

	// 用完context 再去调用
	cancel()
}

func TestContext_WithCancel(t *testing.T) {
	ctx := context.Background()

	ctx, cancel := context.WithCancel(ctx)
	// 用完CTX在关闭
	//defer cancel()

	go func() {
		time.Sleep(time.Second)
		fmt.Println("xxx")
		cancel()
	}()

	// 发送
	<-ctx.Done()
	t.Log("hello cancel: ", ctx.Err())
}

func TestContext_WithDeadline(t *testing.T) {
	ctx := context.Background()

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*3))

	deadline, _ := ctx.Deadline()

	t.Log(deadline)

	defer cancel()

	<-ctx.Done()

	t.Log("hello deadline", ctx.Err())
}

func TestContext_WithTimeout(t *testing.T) {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)

	deadline, _ := ctx.Deadline()

	t.Log(deadline)

	defer cancel()

	<-ctx.Done()

	t.Log("hello deadline", ctx.Err())
}

// 父无法访问子ctx内容
// 逼不得已的情况下，在父ctx中放入map。后续操作都是修改这个map
