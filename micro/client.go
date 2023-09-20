package micro

import (
	"GoWebFrame/micro/register"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
	"time"
)

type ClientOption func(c *Client)

type Client struct {
	insecure bool
	rb       resolver.Builder

	balancer balancer.Builder
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

func ClientWithRegistry(r register.Register, timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.rb = NewRegisterBuilder(r, timeout)
	}
}

func ClientWithPickerBuilder(name string, b base.PickerBuilder) ClientOption {
	return func(client *Client) {
		builder := base.NewBalancerBuilder(name, b, base.Config{HealthCheck: true})
		balancer.Register(builder)
		client.balancer = builder
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
