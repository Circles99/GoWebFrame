package GoWebFrame

import (
	"log"
	"net"
	"net/http"
	"strconv"
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

func NewHttpServer() *HttpServer {
	return &HttpServer{
		router: NewRouter(),
	}
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

	if n.node != nil {
		// 赋值给ctx上下文
		ctx.PathParams = n.patchParams
		ctx.MatchedRoute = n.node.route
	}

	// 最后一个应该是执行用户代码
	var root HandleFunc = func(ctx *Context) {
		if !ok || n.node == nil || n.node.handler == nil {
			ctx.RespStatusCode = 404
			return
		}
		// 执行业务逻辑
		n.node.handler(ctx)
	}
	// n.Mdls = [1m, 2m]
	// 2M(handler) -> 1m(2m)
	// m(1m)   m 是flashResp
	//  m  next => 1m  next -> 2m  next-> handler  return ====>   2m  return-> 1m  return-> m

	// 从后往前组装middleware
	for i := len(n.Mdls) - 1; i >= 0; i-- {
		// 把root直接当参数传入
		root = n.Mdls[i](root)
	}

	// 第一个应该是回写响应的
	// 因为它在调用next之后才回写响应，
	// 所以实际上 flashResp 是最后一个步骤
	var m Middleware = func(next HandleFunc) HandleFunc {
		return func(ctx *Context) {
			next(ctx)
			h.flashResp(ctx)
		}
	}

	// 把返回状态加在最后
	root = m(root)
	// 开始执行中间件
	root(ctx)
}

// Use 加载中间件
func (h *HttpServer) Use(method string, path string, mdls ...Middleware) {
	h.addRouter(method, path, nil, mdls...)
}

func (h *HttpServer) Get(path string, handler HandleFunc, mdls ...Middleware) {
	h.addRouter(http.MethodGet, path, handler, mdls...)
}

func (h *HttpServer) Post(path string, handler HandleFunc, mdls ...Middleware) {
	h.addRouter(http.MethodPost, path, handler, mdls...)
}

func (h *HttpServer) Delete(path string, handler HandleFunc, mdls ...Middleware) {
	h.addRouter(http.MethodDelete, path, handler, mdls...)
}

func (h *HttpServer) Put(path string, handler HandleFunc, mdls ...Middleware) {
	h.addRouter(http.MethodPut, path, handler, mdls...)
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

func (h *HttpServer) flashResp(ctx *Context) {
	if ctx.RespStatusCode > 0 {
		ctx.Resp.WriteHeader(ctx.RespStatusCode)
	}
	ctx.Resp.Header().Set("Content-Length", strconv.Itoa(len(ctx.RespData)))
	_, err := ctx.Resp.Write(ctx.RespData)
	if err != nil {
		// s.log.Fatalln("回写响应失败", err)
		log.Fatalf("回写响应失败", err)
	}
}
