package grpc_resolver

import "google.golang.org/grpc/resolver"

type Builder struct {
}

func (b *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	r := Resolver{
		cc: cc,
	}

	// 需要立刻解析一次，不然client不知道怎么连接
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

func (b *Builder) Scheme() string {
	return "register"
}

type Resolver struct {
	cc resolver.ClientConn
}

func (r Resolver) ResolveNow(options resolver.ResolveNowOptions) {
	// 固定写死ip 和 端口
	err := r.cc.UpdateState(resolver.State{
		Addresses: []resolver.Address{
			{
				Addr: "localhost:8081",
			},
		},
	})
	if err != nil {
		// 没啥用，report之后grpc还是会调用几次resolveNow
		r.cc.ReportError(err)
	}
}

func (r Resolver) Close() {
	//TODO implement me
	panic("implement me")
}
