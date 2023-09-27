package _broadcast

import (
	"GoWebFrame/micro/register"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"reflect"
	"sync"
)

type ClusterBuilder struct {
	registry    register.Register
	service     string
	dialOptions []grpc.DialOption
}

func (b ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	//method = user.userService/getById
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		ok, ch := isBrodCast(ctx)
		if !ok {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		defer func() {
			close(ch)
		}()

		instances, err := b.registry.ListenServices(ctx, b.service)
		if err != nil {
			return err
		}

		var wg sync.WaitGroup
		tye := reflect.TypeOf(reply).Elem()
		wg.Add(len(instances))
		for _, ins := range instances {

			addr := ins.Address
			go func() {
				insCC, er := grpc.Dial(addr, b.dialOptions...)
				if er != nil {
					ch <- Resp{
						Err:   er,
						Reply: reply,
					}
					wg.Done()
					return
				}

				newReply := reflect.New(tye).Interface()

				// 对每一个节点进行调用
				err = invoker(ctx, method, req, newReply, insCC, opts...)
				// 如果没有人接收就会堵住
				select {
				case <-ctx.Done():
					err = fmt.Errorf("响应没有人接受， %w", ctx.Err())
				case ch <- Resp{Err: er, Reply: newReply}:

				}

				//ch <- Resp{
				//	Err:   er,
				//	Reply: newReply,
				//}
				wg.Done()

			}()
		}
		wg.Wait()
		return nil
	}
}

func useBrodCast(ctx context.Context) (context.Context, <-chan Resp) {
	ch := make(chan Resp)
	return context.WithValue(ctx, broadcastKey{}, ch), ch
}

type broadcastKey struct {
}

func isBrodCast(ctx context.Context) (bool, chan Resp) {
	val, ok := ctx.Value(broadcastKey{}).(chan Resp)
	return ok, val
}

type Resp struct {
	Err   error
	Reply any
}
