package recovery

import (
	"GoWebFrame"
	"log"
	"testing"
)

func TestMiddlewareBuilder_Builder(t *testing.T) {

	h := GoWebFrame.NewHttpServer()

	h.Use("GET", "/user", NewMiddlewareBuilder(500, "ddddd", func(ctx *GoWebFrame.Context) {
		log.Println(ctx.Req.URL.Path)
	}).Builder())
	h.Get("/user", func(c *GoWebFrame.Context) {
		panic("panic")
	})

	h.Start(":8082")
}
