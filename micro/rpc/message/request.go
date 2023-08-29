package message

import (
	"bytes"
	"encoding/binary"
)

type Request struct {
	//uint32=4个字节 一个和字节8个bit
	// 消息长度
	HeadLength uint32
	// 请求体长度
	BodyLength uint32
	// 消息ID
	RequestID uint32
	// 版本 一个字节
	Version uint8
	// 压缩算法
	Compresser uint8
	// 序列化协议
	Serializer uint8

	ServiceName string
	MethodName  string
	// 扩展字段，用于传递元数据
	Meta map[string]string
	//协议体
	Data []byte
}

func EncodeReq(req *Request) []byte {
	// 预期长度 = 头部长度+消息体长度
	bs := make([]byte, req.HeadLength+req.BodyLength)

	//binary 是为了转码的引入
	// 头4个字节写入头部长度
	binary.BigEndian.PutUint32(bs[:4], req.HeadLength)
	// 写入body长度
	binary.BigEndian.PutUint32(bs[4:8], req.BodyLength)
	// 写入RequestId
	binary.BigEndian.PutUint32(bs[8:12], req.RequestID)
	// 写入Version,Compresser,Serializer 只有一个字节,没有编码问题
	bs[12] = req.Version
	bs[13] = req.Compresser
	bs[14] = req.Serializer

	//下标为15开始放，放到ServiceName长度
	cur := bs[15:]
	copy(cur, req.ServiceName)
	cur = cur[len(req.ServiceName):]
	cur[0] = '\n'
	cur = cur[1:]
	copy(cur, req.MethodName)

	// 元数据部分
	cur = cur[len(req.MethodName):]
	cur[0] = '\n'
	cur = cur[1:]
	for key, value := range req.Meta {
		copy(cur, key)
		cur = cur[len(key):]
		cur[0] = '\r'
		cur = cur[1:]
		copy(cur, value)
		cur = cur[len(value):]
		cur[0] = '\n'
		cur = cur[1:]
	}

	copy(cur, req.Data)

	return bs
}

func DecodeReq(data []byte) *Request {
	req := &Request{}

	// 头4个字节是头部长度
	req.HeadLength = binary.BigEndian.Uint32(data[:4])

	// 获取下一个4个字节获取body长度
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])

	// 获取下一个4个字节获取requestId
	req.RequestID = binary.BigEndian.Uint32(data[8:12])
	// 获取version, Compresser, Serializer
	req.Version = data[12]
	req.Compresser = data[13]
	req.Serializer = data[14]

	// body 和 header切割开
	header := data[15:req.HeadLength]

	// 近似于
	//User-service
	//GetById
	index := bytes.IndexByte(header, '\n')
	// 引入分隔符，切分sericeName 和 methodName
	req.ServiceName = string(header[:index])
	// index 所在分割符本身 所以加1
	header = header[index+1:]

	index = bytes.IndexByte(header, '\n')
	req.MethodName = string(header[:index])
	// 跳过分隔符
	header = header[index+1:]
	// 剩下的是meta解析

	index = bytes.IndexByte(header, '\n')

	if index != -1 {
		meta := make(map[string]string, 4)
		// -1 代表没找到index
		for index != -1 {
			pair := header[:index]
			// /r的位置
			pairIndex := bytes.IndexByte(pair, '\r')
			key := string(pair[:pairIndex])
			value := string(pair[pairIndex+1:])
			meta[key] = value
			header = header[index+1:]
			index = bytes.IndexByte(header, '\n')
		}
		req.Meta = meta
	}

	if req.BodyLength != 0 {
		// data从req.HeadLength之后都是body
		req.Data = data[req.HeadLength:]
	}

	return req

}
