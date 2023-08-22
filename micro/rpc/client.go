package rpc

import (
	"context"
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

	req := EncodeMsg(data)
	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	return ReadMsg(conn)
}
