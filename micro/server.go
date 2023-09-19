package micro

import (
	"GoWebFrame/micro/register"
	"context"
	"google.golang.org/grpc"
	"net"
	"time"
)

type ServerOptions func(server *Server)

type Server struct {
	name            string
	addr            string
	register        register.Register
	registerTimeout time.Duration
	*grpc.Server
	listener net.Listener
	weight   uint32
	group    string
}

func NewServer(name string, opts ...ServerOptions) (*Server, error) {
	res := &Server{
		name:            name,
		Server:          grpc.NewServer(),
		registerTimeout: time.Second * 10,
	}

	for _, opt := range opts {
		opt(res)
	}

	return res, nil
}

func ServerWithWeight(weight uint32) ServerOptions {
	return func(server *Server) {
		server.weight = weight
	}
}

// Start 当用户调用这个方法的时候，就是服务已经准备好
func (s Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.listener = listener
	// 要开始注册了

	if s.register != nil {
		//在这里注册
		ctx, cancel := context.WithTimeout(context.Background(), s.registerTimeout)
		defer cancel()
		err = s.register.Register(ctx, register.ServiceInstance{
			Name: s.name,
			// 在容器中不能够使用
			// 容器外的地址：在start的时候传进来，或者在环境变量中获取
			Address: listener.Addr().String(),
			Group:   s.group,
		})
		if err != nil {
			return err
		}
		//// 已经注册成功了
		//defer func() {
		//	// start返回了，代表服务器退出了
		//	//_ = s.register.Close()
		//	//s.register.UnRegister()
		//}()
	}

	err = s.Serve(s.listener)

	return err
}

func (s Server) Close() error {

	if s.register != nil {
		err := s.register.Close()
		if err != nil {
			return err
		}
	}

	// 这里就能关所有
	s.GracefulStop()

	//err := s.listener.Close()
	return nil
}

func ServerWithRegister(r register.Register) ServerOptions {
	return func(server *Server) {
		server.register = r
	}
}

func ServerWithGroup(group string) ServerOptions {
	return func(server *Server) {
		server.group = group
	}
}
