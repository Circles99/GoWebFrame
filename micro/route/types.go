package route

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// 返回值 true：留下来，false：丢弃
type Filter func(info balancer.PickInfo, addr resolver.Address) bool

type GroupFilter struct {
	Group string
}

func (f GroupFilter) Build() Filter {
	return func(info balancer.PickInfo, addr resolver.Address) bool {
		target := addr.Attributes.Value("group").(string)

		input := info.Ctx.Value("group").(string)
		return target == input
	}
}
