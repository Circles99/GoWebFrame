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

	r := NewRouter()

	var mockHandleFunc HandleFunc = func(c *Context) {}

	for _, s := range testRouter {
		r.addRouter(s.method, s.path, mockHandleFunc)
	}

	wantRouter := &router{
		trees: map[string]*node{
			http.MethodGet: &node{
				path: "/",
				children: map[string]*node{
					"user": {
						path:     "home",
						children: nil,
						handler:  mockHandleFunc,
					},
					"order": {path: "order", children: map[string]*node{
						"detail": {path: "detail", handler: mockHandleFunc},
					}},
				},
				handler: nil,
			},
			http.MethodPost: {
				path: "/",
				children: map[string]*node{
					"order": {path: "order", children: map[string]*node{
						"create": {path: "create", handler: mockHandleFunc},
					}},
					"login": {path: "login", handler: mockHandleFunc},
				},
			},
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
	if len(n.children) != len(y.children) {
		return fmt.Sprintf("子节点数量不相同"), false
	}

	if y.path != n.path {
		return fmt.Sprintf("节点路径不匹配"), false
	}

	// 对比handler是否相同

	nHandler := reflect.ValueOf(n.children)
	yHandler := reflect.ValueOf(y.children)
	if nHandler != yHandler {
		return fmt.Sprintf("handler 不一致"), false
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
