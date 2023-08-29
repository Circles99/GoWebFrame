package message

import (
	"encoding/binary"
)

type Response struct {
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
	// 错误
	Error []byte

	Data []byte
}

func (resp *Response) CalculateHeaderLength() {
	// 15是默认计算的 HeadLength + BodyLength + RequestID + Version + Compresser + Serializer
	// 中间的1是为了分隔符留下的

	resp.HeadLength = 15 + uint32(len(resp.Error))
}

func (resp *Response) CalculateBodyLength() {
	resp.BodyLength = uint32(len(resp.Data))
}

func EncodeResp(resp *Response) []byte {
	// 预期长度 = 头部长度+消息体长度
	bs := make([]byte, resp.HeadLength+resp.BodyLength)

	//binary 是为了转码的引入
	// 头4个字节写入头部长度
	binary.BigEndian.PutUint32(bs[:4], resp.HeadLength)
	// 写入body长度
	binary.BigEndian.PutUint32(bs[4:8], resp.BodyLength)
	// 写入RequestId
	binary.BigEndian.PutUint32(bs[8:12], resp.RequestID)
	// 写入Version,Compresser,Serializer 只有一个字节,没有编码问题
	bs[12] = resp.Version
	bs[13] = resp.Compresser
	bs[14] = resp.Serializer

	//下标为15开始放，放到ServiceName长度
	cur := bs[15:]

	copy(cur, resp.Error)
	cur = cur[len(resp.Error):]

	copy(cur, resp.Data)

	return bs
}

func DecodeResp(data []byte) *Response {
	resp := &Response{}

	// 头4个字节是头部长度
	resp.HeadLength = binary.BigEndian.Uint32(data[:4])

	// 获取下一个4个字节获取body长度
	resp.BodyLength = binary.BigEndian.Uint32(data[4:8])

	// 获取下一个4个字节获取requestId
	resp.RequestID = binary.BigEndian.Uint32(data[8:12])
	// 获取version, Compresser, Serializer
	resp.Version = data[12]
	resp.Compresser = data[13]
	resp.Serializer = data[14]

	// 大于15 代表有error
	if resp.HeadLength > 15 {
		resp.Error = data[15:resp.HeadLength]
	}

	if resp.BodyLength != 0 {
		resp.Data = data[resp.HeadLength:]
	}

	return resp

}
