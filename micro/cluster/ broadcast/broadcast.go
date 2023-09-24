package _broadcast

import (
	"GoWebFrame/micro/register"
	"context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ClusterBuilder struct {
	registry    register.Register
	service     string
	dialOptions []grpc.DialOption
}

func (b ClusterBuilder) BuildUnaryInterceptor() grpc.UnaryClientInterceptor {
	//method = user.userService/getById
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !isBrodCast(ctx) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		instances, err := b.registry.ListenServices(ctx, b.service)
		if err != nil {
			return err
		}

		var eg errgroup.Group

		for _, ins := range instances {
			addr := ins.Address
			eg.Go(func() error {
				insCC, er := grpc.Dial(addr, b.dialOptions...)
				if er != nil {
					return err
				}
				// 对每一个节点进行调用
				return invoker(ctx, method, req, reply, insCC, opts...)

			})
		}

		return eg.Wait()
	}
}

func useBrodCast(ctx context.Context) context.Context {
	return context.WithValue(ctx, broadcastKey{}, true)
}

type broadcastKey struct {
}

func isBrodCast(ctx context.Context) bool {
	val, ok := ctx.Value(broadcastKey{}).(bool)
	return ok && val
}
