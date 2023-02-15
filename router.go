package GoWebFrame

import (
	"fmt"
	"strings"
)

// router
// @Description: 路由
// @author liujiming
// @date 2023-01-31 15:58:32
type router struct {
	trees map[string]*node
}

// node
// @Description: 节点
// @author liujiming
// @date 2023-01-31 15:57:04
type node struct {
	path          string           //路径
	children      map[string]*node // 子节点
	handler       HandleFunc
	adaptiveChild *node // 模糊匹配
	paramsChild   *node // 参数匹配
}

type matchInfo struct {
	node        *node
	patchParams map[string]string
}

func NewRouter() *router {
	return &router{
		trees: map[string]*node{},
	}
}

func (r *router) addRouter(method, path string, handleFunc HandleFunc) {
	if path == "" {
		// path不能等于空
		panic("path不可等于空")
	}

	if path[0] != '/' {
		panic("web: 路由必须以 / 开头")
	}

	if path != "/" && path[len(path)-1] == '/' {
		panic("web: 路由不能以 / 结尾")
	}

	root, ok := r.trees[method]
	if !ok {
		// 1： 先建立tress根节点
		root = &node{path: "/"}
		r.trees[method] = root
	}

	// 适配只有一个/的情况
	if path == "/" {
		if root.handler != nil {
			panic("路由冲突[/]")
		}
		root.handler = handleFunc
		return
	}

	segs := strings.Split(path[1:], "/")
	// 切分path, 因为第一个是/，前面已经创建了根节点，所以从之后开始
	for _, s := range segs {
		if s == "" {
			panic(fmt.Sprint("非法路由"))
		}

		// 每次循环覆盖一个新的子节点
		root = root.childOrCreate(s)
	}

	// 多次创建同一个路由，报错
	if root.handler != nil {
		panic(fmt.Sprintf("web: 路由冲突[%s]", path))
	}
	root.handler = handleFunc

}

// findRoute 查找路由
func (r *router) findRoute(method string, path string) (*matchInfo, bool) {
	root, ok := r.trees[method]
	if !ok {
		return nil, false
	}
	// 根节点直接返回
	if path == "/" {
		return &matchInfo{node: root}, true
	}

	var pathParams map[string]string

	for _, s := range strings.Split(strings.Trim(path, "/"), "/") {

		// 一直往下找，找到并且重新赋值往下
		child, paramChild, ok := root.childOf(s)
		if !ok {
			return nil, false
		}

		// 命中了路径参数
		if paramChild {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			//  path是：id, 所以获取第一位之后的
			pathParams[child.path[1:]] = s
		}

	}
	return &matchInfo{
		node:        root,
		patchParams: pathParams,
	}, true

}

// childOf
// @Description:
// @receiver n
// @param path
// @return *node 子节点
// @return bool 标记是否是路径参数
// @return bool 标记是否命中
func (n *node) childOf(path string) (*node, bool, bool) {

	if n.children == nil {
		if n.paramsChild != nil {
			return n.paramsChild, true, true
		}

		// 下级为nil， 判断是否是通配符匹配
		return n.adaptiveChild, false, n.adaptiveChild != nil
	}

	root, ok := n.children[path]
	if !ok {
		// 当没有子节点时
		// 判断是否是通配符或者参数匹配

		if n.paramsChild != nil {
			return n.paramsChild, true, true
		}
		return n.adaptiveChild, false, n.adaptiveChild != nil
	}

	return root, false, ok
}

// childOrCreate 子节点创建
func (n *node) childOrCreate(path string) *node {

	// 不允许同时注册参数匹配或者通配符匹配路由

	// 参数匹配
	if path[0] == ':' {
		if n.adaptiveChild != nil {
			panic("web: 不允许同时注册路径参数匹配和通配符匹配, 已有通配符匹配")
		}

		if n.paramsChild == nil {
			n.paramsChild = &node{path: path}
		}

		return n.paramsChild
	}

	// 通配符匹配
	if path == "*" {
		if n.paramsChild != nil {
			panic("web: 不允许同时注册路径参数匹配和通配符匹配, 已有路径参数")
		}

		if n.adaptiveChild == nil {
			n.adaptiveChild = &node{
				path: path,
			}
		}
		return n.adaptiveChild
	}

	// 当没有子节点时，make一个新的
	if n.children == nil {
		n.children = make(map[string]*node)
	}

	// 获取是否存在这个子节点
	// 没有则插入一个新的
	child, ok := n.children[path]
	if !ok {
		child = &node{path: path}
		n.children[path] = child
	}
	// 返回当前子节点
	return child
}
