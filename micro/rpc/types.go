package rpc

import "context"

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *Request) (*Response, error)
}

type Response struct {
	Data []byte
}
