package round_robin

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"strconv"
	"sync"
	"sync/atomic"
)

type WeightBalancer struct {
	connections []*weightConn
	mutex       sync.Mutex
}

func (w *WeightBalancer) Pick(info balancer.PickInfo) (balancer.PickResult, error) {

	if len(w.connections) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var totalWeight uint32
	var res *weightConn
	w.mutex.Lock()
	defer w.mutex.Unlock()
	for _, c := range w.connections {
		totalWeight = totalWeight + c.efficientWeight
		c.currentWeight = c.currentWeight + c.efficientWeight
		if res == nil || res.currentWeight < c.currentWeight {
			res = c
		}
	}

	res.currentWeight = res.currentWeight - totalWeight
	return balancer.PickResult{
		SubConn: res.c,
		Done: func(info balancer.DoneInfo) {

			// 最简单的就是加个锁
			//w.mutex.Lock()
			//if info.Err == nil && res.efficientWeight == 0 {
			//	return
			//}
			//// err不为nil，当int32达到满了，就会变成0 不能在往上加了
			//if info.Err == nil && res.efficientWeight == math.MaxUint32 {
			//	return
			//}
			//
			//if info.Err != nil {
			//	res.efficientWeight--
			//} else {
			//	res.efficientWeight++
			//}
			//
			//w.mutex.Unlock()

			for {

				// 因为是并发的，需要原子操作
				weight := atomic.LoadUint32(&res.efficientWeight)
				if info.Err == nil && weight == 0 {
					return
				}

				// err不为nil，当int32达到满了，就会变成0 不能在往上加了
				if info.Err == nil && weight == math.MaxUint32 {
					return
				}

				newWeight := weight

				if info.Err != nil {
					newWeight--
				} else {
					newWeight++
				}
				// 确保没人和我抢
				if atomic.CompareAndSwapUint32(&(res.efficientWeight), weight, newWeight) {
					return
				}
			}
		},
	}, nil

}

type WeightBalancerBuilder struct {
}

func (w *WeightBalancerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {

	cs := make([]*weightConn, 0, len(info.ReadySCs))

	for sub, subInfo := range info.ReadySCs {
		weightStr := subInfo.Address.Attributes.Value("weight").(string)

		weight, err := strconv.ParseInt(weightStr, 10, 64)
		if err != nil {
			panic(err)
		}

		cs = append(cs, &weightConn{
			c:               sub,
			weight:          uint32(weight),
			currentWeight:   uint32(weight),
			efficientWeight: uint32(weight),
		})
	}

	return &WeightBalancer{
		connections: cs,
	}
}

type weightConn struct {
	c               balancer.SubConn
	weight          uint32
	currentWeight   uint32
	efficientWeight uint32
}
