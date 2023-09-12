package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"sync/atomic"
)

type Balancer struct {
	index       int32
	connections []balancer.SubConn // 维持一个可轮训的connection
	length      int32
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {

	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	// 此处有线程安全的问题
	// 直接给他加1
	idx := atomic.AddInt32(&b.index, 1)

	// 取余， length 和 connections 初始化之后没人修改， 不需要用原子操作
	c := b.connections[idx%b.length]

	return balancer.PickResult{
		SubConn: c, //对一个实例的连接池的抽象
		Done: func(info balancer.DoneInfo) {
			// 相应回来的时候回调这个方法
		},
	}, nil
}

type Builder struct {
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {

	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		connections = append(connections, c)
	}
	return &Balancer{
		index:       -1,
		connections: connections,
		length:      int32(len(info.ReadySCs)),
	}
}
