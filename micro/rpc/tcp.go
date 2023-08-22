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
	length := binary.BigEndian.Uint64(lenBs)

	data := make([]byte, length)

	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

//func EncodeMsg(data []byte)
