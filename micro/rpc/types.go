package rpc

import "context"

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *Request) (*Response, error)
}

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

type Response struct {
	// 消息长度
	HeadLength uint32
	// 请求体长度
	BodyLength uint32
	// 消息ID
	MessageID uint32
	// 版本 一个字节
	Version uint8
	// 压缩算法
	Compresser uint8
	// 序列化协议
	Serializer uint8
	// 错误
	Error []byte

	Data []byte
}
