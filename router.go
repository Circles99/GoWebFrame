package GoWebFrame

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

}
