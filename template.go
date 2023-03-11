package GoWebFrame

import (
	"bytes"
	"context"
	"html/template"
)

type TemplateInterface interface {
	// data 渲染页面数据
	Render(ctx context.Context, tplName string, data any) ([]byte, error)
}

type TemplateEngine struct {
	T *template.Template
}

func (t TemplateEngine) Render(ctx context.Context, tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := t.T.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err
}
