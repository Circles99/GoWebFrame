package random

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math/rand"
)

type Balancer struct {
	connections []balancer.SubConn
}

func (b Balancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {

	if len(b.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	idx := rand.Intn(len(b.connections))
	return balancer.PickResult{
		SubConn: b.connections[idx],
		Done: func(info balancer.DoneInfo) {

		},
	}, nil
}

type BalancerBuilder struct {
}

func (b BalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	connections := make([]balancer.SubConn, 0, len(info.ReadySCs))
	for c := range info.ReadySCs {
		connections = append(connections, c)
	}
	return &Balancer{
		connections: connections,
	}
}
