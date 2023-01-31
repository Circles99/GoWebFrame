package GoWebFrame

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestRouter_addRouter(t *testing.T) {
	
	testRouters := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/user/home",
		},
		{
			method: http.MethodGet,
			path:   "/",
		},
		{
			method: http.MethodGet,
			path:   "user",
		},
		{
			method: http.MethodGet,
			path:   "//user//adsadsa/ds",
		},
	}

	mockHandler := func(ctx *Context) {}
	r := NewRouter()

	for _, tt := range testRouters {
		r.addRouter(tt.method, tt.path, mockHandler)
	}

	// 断言两者相等
	wantRouter := &Router{
		trees: map[string]*node{},
	}

	msg, ok := wantRouter.equal(r)
	assert.True(t, ok, msg)
}

func (r *Router) equal(y * Router) (string, bool) {
	for k, v := range r.trees {
		dts, ok := y.trees[k]
		if !ok {
			return fmt.Sprint("找不到对应的HTTP method"), false
		}
		msg, ok := v.equal(dts)
		if !ok {
			return msg, false
		}
	}
	return "", true
}


func (n *node) equal(y *node) (string, bool) {
	if n.path != y.path {
		return fmt.Sprintf("节点路径不匹配"), false
	}

	if len(n.children) != len(y.children) {
		return fmt.Sprintf("子节点数量不相等"), false
	}


	// 对比handler不相等
	nHandler := reflect.ValueOf(n.handler)
	yHandler := reflect.ValueOf(y.handler)
	if nHandler != yHandler {
		return fmt.Sprintf("handler 不相等"), false
	}

	//查询子节点是否相等
	for path, c := range n.children {
		dts, ok := y.children[path]
		if !ok {
			return fmt.Sprintf("子节点 %s 不存在",path), false
		}

		msg, ok := c.equal(dts)
		if !ok {
			return msg, false
		}
	}
	return "", true
}
