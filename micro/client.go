package micro

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

type ClientOption func(c *Client)

type Client struct {
	insecure bool
	rb       resolver.Builder
}

func NewClient(opts ...ClientOption) (*Client, error) {
	res := &Client{}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func ClientRb(rb resolver.Builder) ClientOption {
	return func(c *Client) {
		c.rb = rb
	}
}

func (c Client) Dail(ctx context.Context, service string) (*grpc.ClientConn, error) {

	opts := []grpc.DialOption{grpc.WithResolvers(c.rb)}

	if c.insecure {
		opts = append(opts, grpc.WithInsecure())
	}

	cc, err := grpc.DialContext(ctx, fmt.Sprintf("register:///%s", service), opts...)

	return cc, err
}
