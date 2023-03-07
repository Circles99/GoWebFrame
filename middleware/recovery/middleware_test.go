package recovery

import (
	"GoWebFrame"
	"log"
	"testing"
)

func TestMiddlewareBuilder_Builder(t *testing.T) {

	h := GoWebFrame.NewHttpServer()
	h.Get("/user", func(c *GoWebFrame.Context) {
		panic("panic")
	}, NewMiddlewareBuilder(500, "ddddd", func(ctx *GoWebFrame.Context) {
		log.Println(ctx.Req.URL.Path)
	}).Builder())

	h.Start(":8082")
}
