package net

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const numOfLengthBytes = 8

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
	// 连接服务端口
	conn, err := net.DialTimeout(c.network, c.addr, time.Second)
	if err != nil {
		return "", err
	}

	// 退出的时候关闭连接
	defer func() {
		_ = conn.Close()
	}()

	// 获取send数据大小
	reqLen := len(data)
	// 我要在这构建相应数据

	// 这里和服务端，在大送到服务端时，前8个直接就已经是告知整个数据大小

	// 这里+8是需要把数据长度写入
	req := make([]byte, reqLen+numOfLengthBytes)

	// 第一步 把长度写进去前8个字节
	// 编码写入到前8个字节
	binary.BigEndian.PutUint64(req[:numOfLengthBytes], uint64(reqLen))
	// 第二步 把数据写入
	copy(req[8:], data)

	_, err = conn.Write(req)
	if err != nil {
		return "", err
	}

	lenBs := make([]byte, numOfLengthBytes)
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
