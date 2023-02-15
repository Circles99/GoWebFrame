package GoWebFrame

import (
	"fmt"
	"regexp"
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
	adaptiveChild *node          // 模糊匹配
	paramsChild   *node          // 参数匹配
	regxChild     *node          //正则表达式匹配
	regx          *regexp.Regexp //需要匹配的正则
	paramName     string         // 参数名
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
		child, ok := root.childOf(s)
		if !ok {
			return nil, false
		}

		// 命中了路径参数
		if child.paramName != "" {
			if pathParams == nil {
				pathParams = make(map[string]string)
			}
			//  path是：id, 所以获取第一位之后的
			pathParams[child.path[1:]] = s
		}
		root = child

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
func (n *node) childOf(path string) (*node, bool) {

	// 无子节点的时候，查询是否是路径参数
	if n.children == nil {

		if n.regxChild != nil {
			if n.regxChild.regx.Match([]byte(path)) {
				return n.regxChild, true
			}
		}

		if n.paramsChild != nil {
			return n.paramsChild, true
		}

		// 下级为nil， 判断是否是通配符匹配
		return n.adaptiveChild, n.adaptiveChild != nil
	}

	root, ok := n.children[path]

	if !ok {
		// 当没有子节点时
		// 判断是否是通配符或者参数匹配

		if n.paramsChild != nil {
			return n.paramsChild, true
		}
		return n.adaptiveChild, n.adaptiveChild != nil
	}

	return root, ok
}

// childOrCreate 子节点创建
func (n *node) childOrCreate(path string) *node {

	// 不允许同时注册参数匹配或者通配符匹配路由

	// 参数匹配
	if path[0] == ':' {
		paramsName, regx, isReg := n.parseParam(path)

		if isReg {
			// 加入正则节点
			return n.ChildCreateOfRegx(path, regx, paramsName)
		} else {
			// 参数节点
			return n.ChildCreateOfParam(path, paramsName)
		}
	}

	// 通配符匹配
	if path == "*" {
		if n.paramsChild != nil {
			panic("web: 不允许同时注册路径参数匹配和通配符匹配和正则匹配, 已有路径参数")
		}

		if n.regxChild != nil {
			panic("web: 不允许同时注册路径参数匹配和通配符匹配和正则匹配, 已有参数匹配")
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

func (n *node) ChildCreateOfParam(path, paramName string) *node {

	if n.adaptiveChild != nil {
		panic("web: 不允许同时注册路径参数匹配和通配符匹配, 已有通配符匹配")
	}

	if n.regxChild != nil {
		panic("web: 不允许同时注册路径参数匹配和和通配符匹配, 已有通配符匹配")
	}

	if n.paramsChild == nil {
		n.paramsChild = &node{path: path, paramName: paramName}
	}

	return n.paramsChild

}

func (n *node) ChildCreateOfRegx(path, expr, paramName string) *node {
	if n.adaptiveChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有通配符路由。不允许同时注册通配符路由和正则路由 [%s]", path))
	}
	if n.paramsChild != nil {
		panic(fmt.Sprintf("web: 非法路由，已有路径参数路由。不允许同时注册正则路由和参数路由 [%s]", path))
	}

	if n.regxChild != nil {
		//|| paramName != n.
		if n.regxChild.regx.String() != expr {
			panic(fmt.Sprintf("web: 路由冲突，正则路由冲突，已有 %s，新注册 %s", n.regxChild.path, path))
		}
	} else {
		regExpr, err := regexp.Compile(expr)
		if err != nil {
			panic(fmt.Errorf("web: 正则表达式错误 %w", err))
		}
		n.regxChild = &node{
			path:      path,
			regx:      regExpr,
			paramName: paramName,
		}
	}
	return n.regxChild

}

// parseParam
// @Description: 解析参数
// @receiver n
// @param path
// @return string 路径
// @return string 正则
// @return bool 是否正则
func (n *node) parseParam(path string) (string, string, bool) {
	// path 形式为:id(.*)
	path = path[1:]
	// 从第二位开始截取
	segs := strings.SplitN(path, "(", 2)
	if len(segs) == 2 {
		// 获取后面正则那一段
		expr := segs[1]
		// 判断最后一位是不是")"
		if strings.HasSuffix(expr, ")") {
			return segs[0], expr[:len(expr)-1], true
		}
	}

	return path, "", false
}
