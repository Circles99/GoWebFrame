package GoWebFrame

import (
	"html/template"
	"testing"
)

func TestTpl(t *testing.T) {
	tpl, err := template.ParseGlob("testdata/tpls/*.gohtml")
	if err != nil {
		t.Fatal(err)
	}

	s := NewHttpServer(ServerWithTemplateEngine(&TemplateEngine{T: tpl}))
	s.Get("/login", func(c *Context) {
		err = c.Reader("login.gohtml", nil)
		if err != nil {
			t.Fatal(err)
		}
	})

	if err := s.Start(":8082"); err != nil {
		t.Fatal(err)
	}
}
