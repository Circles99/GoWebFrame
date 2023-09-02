package rpc

import (
	"GoWebFrame/micro/rpc/message"
	"GoWebFrame/micro/rpc/serializer"
	"GoWebFrame/micro/rpc/serializer/json"
	"context"
	"errors"
	"github.com/silenceper/pool"
	"net"
	"reflect"
	"strconv"
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

				meta := make(map[string]string, 2)

				// 我确实设置的超时
				if deadline, ok := ctx.Deadline(); ok {
					meta["deadline"] = strconv.FormatInt(deadline.UnixMilli(), 10)
				}

				if isOneway(ctx) {
					meta["one-way"] = "true"
				}

				req := &message.Request{
					ServiceName: service.Name(),
					MethodName:  fieldTyp.Name,
					Meta:        meta,
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
	pool       pool.Pool
}

type ClientOption func(client *Client)

func NewClient(addr string, options ...ClientOption) (*Client, error) {
	config := &pool.Config{
		InitialCap: 5,
		MaxCap:     30,
		MaxIdle:    20,
		Factory: func() (interface{}, error) {
			return net.Dial("tcp", addr)
		},
		Close: func(i interface{}) error {
			return i.(net.Conn).Close()
		},
		IdleTimeout: time.Minute,
	}

	connPool, err := pool.NewChannelPool(config)
	if err != nil {
		return nil, err
	}

	res := &Client{
		pool:       connPool,
		serializer: &json.Serializer{},
	}

	for _, opt := range options {
		opt(res)
	}

	return res, nil
}

func (c Client) Invoke(ctx context.Context, req *message.Request) (*message.Response, error) {

	// 超时控制代码，但无法控制到发送里面
	// 比如已经超时了，但是send返回了
	// 每次进来都需要创建个channel，性能损害
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	ch := make(chan struct{})
	var (
		resp *message.Response
		err  error
	)

	go func() {
		resp, err = c.doInvoke(ctx, req)
		ch <- struct{}{}
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-ch:
		return resp, err
	}
}

func (c *Client) doInvoke(ctx context.Context, req *message.Request) (*message.Response, error) {
	data := message.EncodeReq(req)

	resp, err := c.Send(ctx, data)
	if err != nil {
		return nil, err
	}

	return message.DecodeResp(resp), nil
}

func (c *Client) Send(ctx context.Context, data []byte) ([]byte, error) {
	val, err := c.pool.Get()
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = c.pool.Put(val)
	}()

	conn := val.(net.Conn)
	//req := EncodeMsg(data)
	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	if isOneway(ctx) {
		// 返回一个 error，防止有用户真的去接收结果
		return nil, errors.New("这是 oneway 调用")
	}

	return ReadMsg(conn)
}
