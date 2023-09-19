package hash

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

type ConsistentBalancer struct {
	connections []balancer.SubConn
}

func (b ConsistentBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {

	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	// grpc拿不到请求，无法根据请求特性做负载均衡
	// 但可以放入info的ctx中
	//info.Ctx.Value("xxx")

	//idx := rand.Intn(len(b.connections))
	return balancer.PickResult{
		SubConn: b.connections[0],
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type ConsistentBalancerBuilder struct {
}

func (b ConsistentBalancer) Build(info base.PickerBuildInfo) balancer.Picker {
	// 节点从这来
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		connections = append(connections, c)
	}
	return &Balancer{
		connections: connections,
	}
}
