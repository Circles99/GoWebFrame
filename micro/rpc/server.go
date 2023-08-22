package rpc

import (
	"encoding/json"
	"errors"
	"golang.org/x/net/context"
	"net"
	"reflect"
)

type Server struct {
	services map[string]reflectionStub
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]reflectionStub, 16),
	}
}

func (s Server) RegisterService(service Service) {
	s.services[service.Name()] = reflectionStub{
		s:     service,
		value: reflect.ValueOf(service),
	}
}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		// 一般是端口被占用
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if er := s.handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

// 我们可以认为，一个请求包含两部分
// 1.长度字段：用八个字节表示
// 2. 请求数据
// 响应也是这个规范
func (s *Server) handleConn(conn net.Conn) error {
	for {

		reqBs, err := ReadMsg(conn)
		if err != nil {
			return err
		}

		// 还愿调用信息
		req := &Request{}
		err = json.Unmarshal(reqBs, req)
		if err != nil {
			return err
		}
		resp, err := s.Invoke(context.Background(), req)

		if err != nil {
			// 这个可能是业务error
			return err
		}

		res := EncodeMsg(resp.Data)

		_, err = conn.Write(res)
		if err != nil {
			return err
		}

		//if n != len(respData) {
		//	return errors.New("没写完数据")
		//}
	}
}

func (s *Server) Invoke(ctx context.Context, req *Request) (*Response, error) {

	// 还愿了调用信息，已经知道参数
	// 发起业务调用
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("调用的服务不存在")
	}

	resp, err := service.Invoke(ctx, req.MethodName, req.Arg)
	if err != nil {
		return nil, err
	}
	return &Response{Data: resp}, nil
}

type reflectionStub struct {
	s     Service
	value reflect.Value
}

func (r reflectionStub) Invoke(ctx context.Context, methodName string, data []byte) ([]byte, error) {

	method := r.value.MethodByName(methodName)

	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())

	// 这个arg真正的类型
	//val.Type().In(0)
	// 重新造一个arg真正的类型进行赋值
	inReq := reflect.New(method.Type().In(1).Elem())
	err := json.Unmarshal(data, inReq.Interface())
	if err != nil {
		return nil, err
	}

	in[1] = inReq
	results := method.Call(in)
	// result[0]是返回值
	// result[1]是error
	if results[1].Interface() != nil {
		return nil, results[1].Interface().(error)
	}

	return json.Marshal(results[0].Interface())
}
