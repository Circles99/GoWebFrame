package recovery

import (
	"GoWebFrame"
)

type MiddlewareBuilder struct {
	StatusCode int
	ErrorMsg   string
	LogFunc    func(ctx *GoWebFrame.Context)
}

func NewMiddlewareBuilder(statusCode int, ErrorMsg string, logFunc func(ctx *GoWebFrame.Context)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		StatusCode: statusCode,
		ErrorMsg:   ErrorMsg,
		LogFunc:    logFunc,
	}
}

func (m *MiddlewareBuilder) Builder() GoWebFrame.Middleware {
	return func(next GoWebFrame.HandleFunc) GoWebFrame.HandleFunc {
		return func(ctx *GoWebFrame.Context) {
			defer func() {
				if err := recover(); err != nil {
					ctx.RespStatusCode = m.StatusCode
					ctx.RespData = []byte(m.ErrorMsg)
					m.LogFunc(ctx)
				}
			}()

			// 执行下一个
			next(ctx)
		}
	}
}
