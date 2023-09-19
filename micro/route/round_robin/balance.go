package round_robin

import (
	"GoWebFrame/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"sync/atomic"
)

type Balancer struct {
	index       int32
	connections []subConn // 维持一个可轮训的connection
	length      int32
	filter      route.Filter
}

func (b *Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	candidates := make([]subConn, 0, len(b.connections))

	for _, c := range b.connections {
		if b.filter != nil && !b.filter(info, c.addr) {
			continue
		}
		candidates = append(candidates, c)
	}

	if len(candidates) == 0 {
		// 也可以考虑 筛选完以后没有可用的节点 直接使用默认节点
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	// 此处有线程安全的问题
	// 直接给他加1
	idx := atomic.AddInt32(&b.index, 1)

	// 取余， length 和 connections 初始化之后没人修改， 不需要用原子操作
	c := b.connections[idx%b.length]

	return balancer.PickResult{
		SubConn: c.c, //对一个实例的连接池的抽象
		Done: func(info balancer.DoneInfo) {
			// 相应回来的时候回调这个方法
		},
	}, nil
}

type Builder struct {
	filter route.Filter
}

func (b *Builder) Build(info base.PickerBuildInfo) balancer.Picker {

	connections := make([]subConn, 0, len(info.ReadySCs))
	for c, ci := range info.ReadySCs {
		connections = append(connections, subConn{
			c:    c,
			addr: ci.Address,
		})
	}

	var filter route.Filter = func(info balancer.PickInfo, addr resolver.Address) bool {
		return true
	}

	if b.filter != nil {
		filter = b.filter
	}

	return &Balancer{
		index:       -1,
		connections: connections,
		length:      int32(len(info.ReadySCs)),
		filter:      filter,
	}
}

type subConn struct {
	c    balancer.SubConn
	addr resolver.Address
}
