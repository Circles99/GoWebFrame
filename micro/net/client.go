package net

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func Connect(network, addr string) error {
	conn, err := net.DialTimeout(network, addr, time.Second)
	if err != nil {
		return err
	}

	defer func() {
		_ = conn.Close()
	}()

	for {
		_, err := conn.Write([]byte("hello"))
		if err != nil {
			return err
		}

		res := make([]byte, 128)
		_, err = conn.Read(res)
		if err != nil {
			return err
		}
		fmt.Println(res)
	}
}

type Client struct {
	network string
	addr    string
}

func (c *Client) Send(data string) (string, error) {
	conn, err := net.DialTimeout(c.network, c.addr, time.Second)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = conn.Close()
	}()

	reqLen := len(data)
	// 我要在这构建相应数据

	// 这里+8是因为上面又8个字节
	req := make([]byte, reqLen+8)

	// 第一步 把长度写进去前8个字节
	binary.BigEndian.PutUint64(req[:8], uint64(reqLen))
	// 第二步 把长度写入数据
	copy(req[8:], data)

	_, err = conn.Write(req)
	if err != nil {
		return "", err
	}

	lenBs := make([]byte, 8)
	_, err = conn.Read(lenBs)
	if err != nil {
		return "", err
	}

	// todo 大顶端和小顶端，看编码是高位在第一个字节还是地位在第一个字节，
	// 我响应有多长
	length := binary.BigEndian.Uint64(lenBs)

	respBs := make([]byte, length)

	_, err = conn.Read(respBs)
	if err != nil {
		return "", err
	}

	return string(respBs), nil
}
