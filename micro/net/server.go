package net

import (
	"encoding/binary"
	"errors"
	"net"
)

func server(network, addr string) error {
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
			if er := handleConn(conn); er != nil {
				_ = conn.Close()
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	for {
		bs := make([]byte, 8)
		n, err := conn.Read(bs)
		if err != nil {
			return err
		}

		if n != 8 {
			return errors.New("没读够数据")
		}

		res := handleMsg(bs)
		n, err = conn.Write(res)
		if err != nil {
			return err
		}

		if n != len(res) {
			return errors.New("没写完数据")
		}
	}
}

func handleMsg(req []byte) []byte {
	res := make([]byte, len(req))

	copy(res[:len(req)], req)
	copy(res[len(req):], req)
	res = append(res, req...)
	return res
}

type Server struct {
	network string
	addr    string
}

func NewServer(network, addr string) *Server {
	return &Server{
		network: network,
		addr:    addr,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen(s.network, s.addr)
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
// 先读一部分数据，比如先读8个字节，然后在读后面的
func (s *Server) handleConn(conn net.Conn) error {
	for {
		// lenBs 是长度字段的字节表示
		// 我用8个字节获取到了长度数字
		lenBs := make([]byte, numOfLengthBytes)
		_, err := conn.Read(lenBs)
		if err != nil {
			return err
		}

		// todo 大顶端和小顶端，看编码是高位在第一个字节还是地位在第一个字节，
		// 我消息有多长
		// 解码获取长度
		length := binary.BigEndian.Uint64(lenBs)

		reqBs := make([]byte, length)

		_, err = conn.Read(reqBs)
		if err != nil {
			return err
		}

		respData := handleMsg(reqBs)

		// 响应长度
		respLen := len(respData)
		// 我要在这构建相应数据

		// 这里+8是因为回到客户端，客户端也需要知道数据有多长
		res := make([]byte, respLen+numOfLengthBytes)

		// 第一步 把长度写进去前8个字节
		binary.BigEndian.PutUint64(res[:numOfLengthBytes], uint64(respLen))
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
