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

		respData := handleMsg(reqBs)

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
