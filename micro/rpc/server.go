package rpc

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"golang.org/x/net/context"
	"net"
	"reflect"
)

type Server struct {
	services map[string]Service
}

func NewServer() *Server {
	return &Server{
		services: make(map[string]Service, 16),
	}
}

func (s Server) RegisterService(service Service) {
	s.services[service.Name()] = service
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
		// lenBs 是长度字段的字节表示
		lenBs := make([]byte, 8)
		_, err := conn.Read(lenBs)
		if err != nil {
			return err
		}

		// todo 大顶端和小顶端，看编码是高位在第一个字节还是地位在第一个字节，
		// 我消息有多长
		length := binary.BigEndian.Uint64(lenBs)

		reqBs := make([]byte, length)

		_, err = conn.Read(reqBs)
		if err != nil {
			return err
		}

		respData, err := s.handleMsg(reqBs)
		if err != nil {
			// 这个可能是业务error
			return err
		}

		respLen := len(respData)
		// 我要在这构建相应数据

		// 这里+8是因为上面又8个字节
		res := make([]byte, respLen+8)

		// 第一步 把长度写进去前8个字节
		binary.BigEndian.PutUint64(res[:8], uint64(respLen))
		// 第二步 把长度写入数据
		copy(res[8:], respData)

		_, err = conn.Write(res)
		if err != nil {
			return err
		}

		//if n != len(respData) {
		//	return errors.New("没写完数据")
		//}
	}
}

func (s *Server) handleMsg(reqData []byte) ([]byte, error) {
	// 还愿调用信息
	req := &Request{}
	err := json.Unmarshal(reqData, req)
	if err != nil {
		return nil, err
	}
	// 还愿了调用信息，已经知道参数
	// 发起业务调用
	service, ok := s.services[req.ServiceName]
	if !ok {
		return nil, errors.New("调用的服务不存在")
	}

	// 反射找到方法，并且执行调用
	val := reflect.ValueOf(service)

	method := val.MethodByName(req.MethodName)

	in := make([]reflect.Value, 2)
	in[0] = reflect.ValueOf(context.Background())

	// 这个arg真正的类型
	//val.Type().In(0)
	// 重新造一个arg真正的类型进行赋值
	inReq := reflect.New(method.Type().In(1).Elem())
	err = json.Unmarshal(req.Arg, inReq.Interface())
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
