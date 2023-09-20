package micro

import (
	"GoWebFrame/micro/register"
	"context"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

type grpcBuilder struct {
	r       register.Register
	timeout time.Duration
}

func NewRegisterBuilder(r register.Register, timeout time.Duration) *grpcBuilder {
	return &grpcBuilder{r: r, timeout: timeout}
}

func (b *grpcBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	r := grpcResolver{
		target:  target,
		r:       b.r,
		cc:      cc,
		timeout: b.timeout,
	}

	// 需要立刻解析一次，不然client不知道怎么连接
	r.resolve()
	go r.watch()
	return r, nil
}

func (b *grpcBuilder) Scheme() string {
	return "register"
}

type grpcResolver struct {
	target  resolver.Target
	r       register.Register
	cc      resolver.ClientConn
	timeout time.Duration
	close   chan struct{}
}

func (r grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	// 固定写死ip 和 端口
	//err := r.cc.UpdateState(resolver.State{
	//	Addresses: []resolver.Address{
	//		{
	//			Addr: "localhost:8081",
	//		},
	//	},
	//})
	//if err != nil {
	//	// 没啥用，report之后grpc还是会调用几次resolveNow
	//	r.cc.ReportError(err)
	//}

	r.resolve()
}

func (r grpcResolver) watch() {
	events, err := r.r.Subscribe(r.target.Endpoint)
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	for {
		select {
		case <-events:
			// resolve 都是幂等的
			r.resolve()
		//case event := <-events:
		//	switch event.Type {
		//	case "DELETE":
		//		// 删除已有的节点
		//
		//	}

		case <-r.close:
			return
		}

	}

}

func (r grpcResolver) resolve() {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()

	instance, err := r.r.ListenServices(ctx, r.target.Endpoint)
	if err != nil {
		r.cc.ReportError(err)
		return
	}
	address := make([]resolver.Address, 0, len(instance))

	for _, si := range instance {
		address = append(address, resolver.Address{
			Addr:       si.Address,
			Attributes: attributes.New("weight", si.Weight).WithValue("group", si.Group),
		})
	}

	err = r.cc.UpdateState(resolver.State{
		Addresses: address,
	})
	if err != nil {
		r.cc.ReportError(err)
		return
	}
}

func (r grpcResolver) Close() {
	close(r.close)
}
