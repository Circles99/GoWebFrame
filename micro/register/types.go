package register

import (
	"context"
	"io"
)

type Register interface {
	// 服务端注册
	Register(ctx context.Context, si ServiceInstance) error
	UnRegister(ctx context.Context, si ServiceInstance) error
	// 客户端
	ListenServices(ctx context.Context, serviceName string) ([]ServiceInstance, error)
	Subscribe(serviceName string) (<-chan Event, error)

	io.Closer
}

type ServiceInstance struct {
	Name string
	// Addr 最关键的定位信息
	Address string
	// 下面可以任意加字段，完全取决于你的服务治理要做成什么样子
	Weight uint32

	Group string
}

type Event struct {
	// add delete
	Type string
}
