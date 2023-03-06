package assesslog

import (
	"GoWebFrame"
	"encoding/json"
	"fmt"
)

type MiddlewareBuild struct {
	logFunc func(accessLog string)
}

func (m *MiddlewareBuild) LogFunc(logFunc func(accessLog string)) *MiddlewareBuild {
	m.logFunc = logFunc
	return m
}

func NewBuilder() *MiddlewareBuild {
	return &MiddlewareBuild{
		logFunc: func(accessLog string) {
			fmt.Println(accessLog)
		},
	}
}

type accessLog struct {
	Host       string
	Route      string
	HTTPMethod string `json:"http_method"`
	Path       string
}

func (m MiddlewareBuild) Build() GoWebFrame.Middleware {
	return func(next GoWebFrame.HandleFunc) GoWebFrame.HandleFunc {
		return func(c *GoWebFrame.Context) {
			defer func() {
				l := accessLog{
					Host:       c.Req.Host,
					Route:      c.MatchedRoute,
					HTTPMethod: c.Req.Method,
					Path:       c.Req.URL.Path,
				}
				// 结构自己定义的，不可能出错
				bs, _ := json.Marshal(l)
				m.logFunc(string(bs))
			}()
			next(c)
		}
	}

}
