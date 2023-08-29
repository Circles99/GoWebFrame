package rpc

import (
	"GoWebFrame/micro/rpc/message"
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
		req := message.DecodeReq(reqBs)
		if err != nil {
			return err
		}
		resp, err := s.Invoke(context.Background(), req)

		if err != nil {
			// 这个可能是业务error
			resp.Error = []byte(err.Error())
			return nil
		}

		//res := EncodeMsg(resp.Data)

		resp.CalculateHeaderLength()
		resp.CalculateBodyLength()

		_, err = conn.Write(message.EncodeResp(resp))
		if err != nil {
			return err
		}

		//if n != len(respData) {
		//	return errors.New("没写完数据")
		//}
	}
}

func (s *Server) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {

	// 还愿了调用信息，已经知道参数
	// 发起业务调用
	service, ok := s.services[req.ServiceName]
	resp := &message.Response{
		RequestID:  req.RequestID,
		Version:    req.Version,
		Compresser: req.Compresser,
		Serializer: req.Serializer,
	}

	if !ok {
		return resp, errors.New("调用的服务不存在")
	}
	respData, err := service.Invoke(ctx, req.MethodName, req.Data)

	resp.Data = respData
	if err != nil {
		return resp, err
	}

	return resp, nil
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
		err = results[1].Interface().(error)
	}

	// 用户不管怎么返回 都可以正确处理掉他
	var res []byte
	if results[0].IsNil() {
		return nil, err
	} else {
		var er error
		res, er = json.Marshal(results[0].Interface())
		if er != nil {
			return nil, er
		}
	}

	return res, err
}
