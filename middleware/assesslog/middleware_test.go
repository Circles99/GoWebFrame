package assesslog

import (
	"GoWebFrame"
	"fmt"
	"testing"
)

func TestMiddlewareBuild_Build(t *testing.T) {
	h := GoWebFrame.NewHttpServer()
	b := NewBuilder()

	//r := GoWebFrame.NewRouter()

	//r.AddRouter("GET", "/user/eee", func(c *GoWebFrame.Context) {
	//	fmt.Println("你好")
	//}, b.Build())
	//e, _ := r.FindRoute("GET", "user/eee")
	//fmt.Println(e)
	h.Get("/user/eee", func(c *GoWebFrame.Context) {
		c.RespData = []byte("<Html><h1>你好</h1></Html>")
	}, b.Build())

	//h.Use(http.MethodGet, "/qq", b.Build())

	err := h.Start(":8082")
	if err != nil {
		fmt.Println(err)
	}
}
