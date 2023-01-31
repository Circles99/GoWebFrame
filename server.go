package GoWebFrame

import (
	"net"
	"net/http"
)

// 确保一定实现了接口
var _ Server = &HttpServer{}

type HandleFunc func(c *Context)

type Server interface {
	http.Handler
	// Start 启动服务器
	Start(addr string) error
}

type HttpServer struct {
	*router
}

func (h *HttpServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	ctx := &Context{
		Req:  request,
		Resp: writer,
	}
	h.serve(ctx)
}

func (h *HttpServer) serve(ctx *Context) {
	// 查找路由，执行命中业务逻辑
}

func (h *HttpServer) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	// 可实现用户注册的回调等
	// 执行一些业务前置操作条件
	return http.Serve(l, h)
}
