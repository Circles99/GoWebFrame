package leastactive

import (
	"GoWebFrame/micro/route"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"math"
	"sync/atomic"
)

type Balancer struct {
	connections []*activeConn
	filter      route.Filter
}

func (b Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {

	candidates := make([]*activeConn, 0, len(b.connections))

	for _, c := range b.connections {
		if b.filter != nil && !b.filter(info, c.addr) {
			continue
		}
		candidates = append(candidates, c)
	}

	if len(candidates) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}
	res := &activeConn{
		cnt: math.MaxUint32,
	}
	// 需要加锁或者原子操作，
	for _, c := range candidates {
		if atomic.LoadUint32(&c.cnt) < res.cnt {
			// c 的数更小 选C
			res = c
		}
	}

	// 连接数+1
	atomic.AddUint32(&res.cnt, 1)

	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {
			// 完成之后连接数-1
			atomic.AddUint32(&res.cnt, -1)
		},
	}, nil
}

type BalancerBuilder struct {
	Filter route.Filter
}

func (b BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]*activeConn, 0, len(info.ReadySCs))
	for c, ci := range info.ReadySCs {
		connections = append(connections, &activeConn{c: c, addr: ci.Address})
	}
	return &Balancer{
		connections: connections,
	}
}

type activeConn struct {
	// 正在请求的数量
	cnt  uint32
	c    balancer.SubConn
	addr resolver.Address
}
