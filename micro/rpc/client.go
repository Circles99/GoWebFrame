package rpc

import (
	"GoWebFrame/micro/rpc/message"
	"GoWebFrame/micro/rpc/serializer"
	"GoWebFrame/micro/rpc/serializer/json"
	"context"
	"errors"
	"net"
	"reflect"
	"time"
)

// todo 这里使用的代理模式进行rpc实现
// 不是grpc那种生成文件

// InitClientProxy 要为 GetById之类的函数类型字段赋值
func (c *Client) InitClientProxy(addr string, service Service) error {
	//client := NewClient(addr)
	return setFuncField(service, c, c.serializer)
}

func setFuncField(service Service, p Proxy, s serializer.Serializer) error {
	if service == nil {
		return errors.New("不支持nil")
	}

	val := reflect.ValueOf(service)
	typ := val.Type()
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Pointer {
		return errors.New("只支持指向结构体的一级指针")
	}

	val = val.Elem()
	typ = typ.Elem()

	numField := typ.NumField()
	for i := 0; i < numField; i++ {
		fieldTyp := typ.Field(i)
		fieldVal := val.Field(i)
		// 我要设置值给 getbyId

		if fieldVal.CanSet() {
			// 创建函数
			fnVal := reflect.MakeFunc(fieldTyp.Type, func(args []reflect.Value) (results []reflect.Value) {
				//这个地方才是真正的将本地调用捕捉到的地方

				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())

				// args[0]是context
				ctx := args[0].Interface().(context.Context)
				// args[1]是request
				reqData, err := s.Encode(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,

					//slice.Map[reflect.Value, any](args, func(idx int, src reflect.Value) any {
					//	return src.Interface()
					//}),
					Serializer: s.Code(),
					Data:       reqData,
				}
				req.CalculateHeaderLength()
				req.CalculateBodyLength()
				// 要真的发起调用
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				var serverErr error
				if len(resp.Error) > 0 {
					// 处理业务error
					serverErr = errors.New(string(resp.Error))
				}

				if len(resp.Data) > 0 {
					err = s.Decode(resp.Data, retVal.Interface())
					if err != nil {
						// 返序列化的errir
						return []reflect.Value{retVal, reflect.ValueOf(err)}
					}
				}

				var retErrVal reflect.Value
				if serverErr == nil {
					retErrVal = reflect.Zero(reflect.TypeOf(new(error)).Elem())
				} else {
					retErrVal = reflect.ValueOf(serverErr)
				}

				return []reflect.Value{retVal, retErrVal}
			})

			fieldVal.Set(fnVal)
		}

	}

	return nil
}

type Client struct {
	serializer serializer.Serializer

	addr string
}

type ClientOption func(client *Client)

func NewClient(addr string, options ...ClientOption) *Client {
	res := &Client{
		addr:       addr,
		serializer: &json.Serializer{},
	}

	for _, opt := range options {
		opt(res)
	}

	return res
}

func (c Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {

	data := message.EncodeReq(req)
	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}

	return message.DecodeResp(resp), nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", c.addr, time.Second)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = conn.Close()
	}()

	//req := EncodeMsg(data)
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	return ReadMsg(conn)
}
