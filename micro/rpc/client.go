package rpc

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"reflect"
	"time"
)

// todo 这里使用的代理模式进行rpc实现
// 不是grpc那种生成文件

// InitClientProxy 要为 GetById之类的函数类型字段赋值
func InitClientProxy(addr string, service Service) error {
	client := NewClient(addr)
	return setFuncField(service, client)
}

func setFuncField(service Service, p Proxy) error {
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
				// args[0]是context
				ctx := args[0].Interface().(context.Context)
				// args[1]是request
				retVal := reflect.New(fieldTyp.Type.Out(0).Elem())
				reqData, err := json.Marshal(args[1].Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				req := &Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,

					//slice.Map[reflect.Value, any](args, func(idx int, src reflect.Value) any {
					//	return src.Interface()
					//}),
					Arg: reqData,
				}

				// 要真的发起调用
				resp, err := p.Invoke(ctx, req)
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				err = json.Unmarshal(resp.Data, retVal.Interface())
				if err != nil {
					return []reflect.Value{retVal, reflect.ValueOf(err)}
				}

				return []reflect.Value{retVal, reflect.Zero(reflect.TypeOf(new(error)).Elem())}
			})

			fieldVal.Set(fnVal)
		}

	}

	return nil
}

type Client struct {
	addr string
}

func NewClient(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

func (c Client) Invoke(ctx context.Context, req *Request) (*Response, error) {

	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.Send(data)
	if err != nil {
		return nil, err
	}

	return &Response{
		Data: resp,
	}, nil
}

func (c *Client) Send(data []byte) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", c.addr, time.Second)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	lenBs := make([]byte, 8)
	_, err = conn.Read(lenBs)
	if err != nil {
		return nil, err
	}

	// todo 大顶端和小顶端，看编码是高位在第一个字节还是地位在第一个字节，
	// 我响应有多长
	length := binary.BigEndian.Uint64(lenBs)

	respBs := make([]byte, length)

	_, err = conn.Read(respBs)
	if err != nil {
		return nil, err
	}

	return respBs, nil
}
