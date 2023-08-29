package rpc

import (
	"encoding/binary"
	"net"
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	lenBs := make([]byte, 8)

	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}

	// todo 大顶端和小顶端，看编码是高位在第一个字节还是地位在第一个字节，
	// 我响应有多长
	headerLength := binary.BigEndian.Uint32(lenBs[:4])
	bodyLength := binary.BigEndian.Uint32(lenBs[4:])
	length := headerLength + bodyLength

	data := make([]byte, length)

	_, err = conn.Read(data[8:])
	if err != nil {
		return nil, err
	}
	copy(data[:8], lenBs)
	return data, nil
}

func EncodeMsg(data []byte) []byte {
	reqLen := len(data)
	// 我要在这构建相应数据

	// 这里+8是因为上面又8个字节
	res := make([]byte, reqLen+8)

	// 第一步 把长度写进去前8个字节
	binary.BigEndian.PutUint64(res[:8], uint64(reqLen))
	// 第二步 把长度写入数据
	copy(res[8:], data)
	return res
}
