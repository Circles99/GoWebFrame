package GoWebFrame



/* node 节点 */
type node struct {
	// 节点数
	path string
	// 子path到字节点的映射
	children map[string]*node
	// 实际逻辑
	handler HandleFunc

}

// Router 路由 */
type Router struct {
	trees map[string]*node
}


func NewRouter() *Router {
	return &Router{
		trees: map[string]*node{},
	}
}


// addRouter 增加路由
func(r * Router) addRouter(method, path string, handle HandleFunc) {

}


