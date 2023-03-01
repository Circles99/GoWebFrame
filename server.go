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
	n, ok := h.findRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok || n.node.handler == nil {
		ctx.Resp.WriteHeader(404)
		ctx.Resp.Write([]byte("Not Found"))
		return
	}
	// 赋值给ctx上下文
	ctx.PathParams = n.patchParams
	// 执行业务逻辑
	n.node.handler(ctx)
}

func (h *HttpServer) Get(path string, handler HandleFunc) {
	h.addRouter(http.MethodGet, path, handler)
}

func (h *HttpServer) Post(path string, handler HandleFunc) {
	h.addRouter(http.MethodPost, path, handler)
}

func (h *HttpServer) Delete(path string, handler HandleFunc) {
	h.addRouter(http.MethodDelete, path, handler)
}

func (h *HttpServer) Put(path string, handler HandleFunc) {
	h.addRouter(http.MethodPut, path, handler)
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
