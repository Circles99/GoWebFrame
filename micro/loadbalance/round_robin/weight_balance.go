package round_robin

import (
	"fmt"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
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
		// 可以改成connection中放锁。让锁的粒度更小，但是可能会导致准确性更差
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
		weight := subInfo.Address.Attributes.Value("weight").(uint32)
		//weightStr := subInfo.Address.Attributes.Value("weight").(uint32)
		//
		//weight, err := strconv.ParseInt(weightStr, 10, 64)
		//if err != nil {
		//	panic(err)
		//}

		cs = append(cs, &weightConn{
			c:               sub,
			weight:          weight,
			currentWeight:   weight,
			efficientWeight: weight,
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

// 在这个示例中，我们首先定义了三个节点，每个节点有不同的有效权重（EffectiveWeight）。然后，我们模拟了一系列请求的选择过程，其中每个请求的结果是随机生成的，用 simulateRequest 函数来表示请求成功或失败。
//
// 在每次请求选择节点后，我们根据请求的结果来调整节点的权重，使用 updateNodeWeight 函数。如果请求成功，节点的有效权重会逐渐增加，反之会逐渐减少。
//
// 通过运行这个示例，您可以看到每次请求选择的节点，以及节点权重的变化。这个示例演示了带有动态权重调整的加权轮询负载均衡算法的工作原理。请注意，这个示例中的权重调整是基于随机模拟的请求结果，实际情况下可能会根据实际情况来调整节点的权重
type Node struct {
	Name            string
	Weight          uint32
	CurrentWeight   uint32
	EffectiveWeight uint32
}

func XXX() {
	nodes := []Node{
		{Name: "NodeA", Weight: 3, CurrentWeight: 0, EffectiveWeight: 3},
		{Name: "NodeB", Weight: 2, CurrentWeight: 0, EffectiveWeight: 2},
		{Name: "NodeC", Weight: 1, CurrentWeight: 0, EffectiveWeight: 1},
	}

	totalWeight := uint32(6) // 总有效权重
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 10; i++ {
		node := selectNode(nodes, totalWeight)
		fmt.Printf("选择节点：%s\n", node.Name)
		success := simulateRequest(node)
		updateNodeWeight(node, success)
		fmt.Printf("节点权重：%s(%d)\n", node.Name, node.EffectiveWeight)
	}
}

func selectNode(nodes []Node, totalWeight uint32) Node {
	maxWeightNode := nodes[0]
	maxWeight := uint32(0)

	for _, node := range nodes {
		node.CurrentWeight += node.EffectiveWeight
		if node.CurrentWeight > maxWeight {
			maxWeight = node.CurrentWeight
			maxWeightNode = node
		}
	}

	maxWeightNode.CurrentWeight -= totalWeight
	return maxWeightNode
}

func simulateRequest(node Node) bool {
	// 模拟请求成功率，这里随机生成一个数字，模拟请求成功或失败
	r := rand.Float64()
	return r < 0.8 // 假设请求成功率为 80%
}

func updateNodeWeight(node Node, success bool) {
	if success {
		if node.EffectiveWeight == 0 {
			return
		}
		if node.EffectiveWeight < math.MaxUint32 {
			node.EffectiveWeight++
		}
	} else {
		if node.EffectiveWeight == math.MaxUint32 {
			return
		}
		if node.EffectiveWeight > 0 {
			node.EffectiveWeight--
		}
	}
}
