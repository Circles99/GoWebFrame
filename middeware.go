package GoWebFrame

// Middleware 中间件
type Middleware func(next HandleFunc) HandleFunc
