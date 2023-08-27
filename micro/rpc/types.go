package rpc

import (
	"GoWebFrame/micro/rpc/message"
	"context"
)

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *message.Request) (*message.Response, error)
}
