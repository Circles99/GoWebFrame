package message

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
