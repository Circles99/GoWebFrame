package GoWebFrame

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_AddRouter(t *testing.T) {
	testRouter := []struct {
		path   string
		method string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/order/detail",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := NewRouter()
	for _, s := range testRouter {
		r.addRouter(s.method, s.path, mockHandler)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: {
				path: "/",
				children: map[string]*node{
					"user": {
						path: "user",
						children: map[string]*node{
							"home": {
								path:    "home",
								handler: mockHandler},
						}, handler: mockHandler},
					"order": {path: "order", children: map[string]*node{
						"detail": {path: "detail", handler: mockHandler},
					}},
				},
				handler: mockHandler,
			},
			http.MethodPost: {path: "/", children: map[string]*node{
				"order": {path: "order", children: map[string]*node{
					"create": {path: "create", handler: mockHandler},
				}},
				"login": {path: "login", handler: mockHandler},
			}},
		},
	}

	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)
}

// equal 比较路由
func (r *router) equal(y *router) (string, bool) {
	for k, v := range r.trees {
		dst, ok := y.trees[k]
		if !ok {
			return fmt.Sprintf("未找到对应的HTTP方法"), false
		}

		msg, ok := v.equal(dst)
		if !ok {
			return msg, false
		}

	}
	return "", true
}

// equal 比较节点
func (n *node) equal(y *node) (string, bool) {

	if y == n {
		return "目标节点为nil", false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("子节点数量不相同"), false
	}

	if y.path != n.path {
		return fmt.Sprintf("节点路径不匹配"), false
	}

	if n.adaptiveChild != nil {
		msg, ok := n.adaptiveChild.equal(y.adaptiveChild)
		if !ok {
			return msg, ok
		}
	}

	if n.paramsChild != nil {
		msg, ok := n.paramsChild.equal(y.paramsChild)
		if !ok {
			return msg, ok
		}
	}

	// 对比handler是否相同
	nhv := reflect.ValueOf(n.handler)
	yhv := reflect.ValueOf(y.handler)
	if nhv != yhv {
		return fmt.Sprintf("%s 节点 handler 不相等 x %s, y %s", n.path, nhv.Type().String(), yhv.Type().String()), false
	}

	// 对比子节点路径
	for path, c := range n.children {
		dst, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("子节点 %s 不存在", path), false
		}
		// 继续往下递归获取
		msg, ok := c.equal(dst)
		if !ok {
			return msg, false
		}
	}
	return "", true
}

func TestFindRouter(t *testing.T) {
	testRouter := []struct {
		path   string
		method string
	}{
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "/user",
		},
		{
			method: http.MethodPost,
			path:   "/order/create",
		},
		{
			method: http.MethodPost,
			path:   "/login/:username",
		},
		{
			method: http.MethodGet,
			path:   "/query/:id(.*)",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := NewRouter()
	for _, s := range testRouter {
		r.addRouter(s.method, s.path, mockHandler)
	}

	testCases := []struct {
		name   string
		method string
		path   string
		found  bool
		info   *matchInfo
	}{
		//{
		//	name:   "method not found",
		//	method: http.MethodHead,
		//},
		//{
		//	name:   "path not found",
		//	method: http.MethodGet,
		//	path:   "/abc",
		//},
		//{
		//	name:   "root",
		//	method: http.MethodGet,
		//	path:   "/",
		//	found:  true,
		//	info: &matchInfo{
		//		node: &node{
		//			path:    "/",
		//			handler: mockHandler,
		//		},
		//	},
		//},
		//{
		//	name:   "user",
		//	method: http.MethodGet,
		//	path:   "/user",
		//	found:  true,
		//	info: &matchInfo{
		//		node: &node{
		//			path:    "user",
		//			handler: mockHandler,
		//		},
		//	},
		//},
		//{
		//	name:   "no handler",
		//	method: http.MethodPost,
		//	path:   "/order",
		//	found:  true,
		//	info: &matchInfo{
		//		node: &node{
		//			path: "order",
		//		},
		//	},
		//},
		//{
		//	name:   "two layer",
		//	method: http.MethodPost,
		//	path:   "/order/create",
		//	found:  true,
		//	info: &matchInfo{
		//		node: &node{
		//			path:    "create",
		//			handler: mockHandler,
		//		},
		//	},
		//},
		//{
		//	name:   "login username",
		//	method: http.MethodPost,
		//	path:   "/login/ljm",
		//	found:  true,
		//	info: &matchInfo{
		//		node: &node{
		//			path:    ":username",
		//			handler: mockHandler,
		//		},
		//	},
		//},
		{
			name:   "query id",
			method: http.MethodGet,
			path:   "/query/西瓜西瓜",
			found:  true,
			info: &matchInfo{
				node: &node{
					path:    ":id(.*)",
					handler: mockHandler,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, found := r.findRoute(tc.method, tc.path)
			assert.Equal(t, tc.found, found)
			if !found {
				return
			}
			fmt.Println(n.patchParams)
			wantVal := reflect.ValueOf(tc.info.node.handler)
			nVal := reflect.ValueOf(n.node.handler)
			assert.Equal(t, wantVal, nVal)
		})
	}
}
