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
	path     string           //路径
	children map[string]*node // 子节点
	handler  HandleFunc
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

	root, ok := r.trees[method]
	if !ok {
		// 1： 先建立tress根节点
		root = &node{path: "/"}
		r.trees[method] = root

	}

	// 切分path, 因为第一个是/，前面已经创建了根节点，所以从之后开始
	for _, s := range strings.Split(path[1:], "/") {
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

// childOrCreate 子节点创建
func (n *node) childOrCreate(path string) *node {

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
