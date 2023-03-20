package GoWebFrame

import (
	"bytes"
	"html/template"
	"mime/multipart"
	"path"
	"testing"
)

func TestFileDownload_Handler(t *testing.T) {

	s := NewHttpServer()
	s.Get("/download", (&FileDownload{Dir: "./testdata/download"}).Handler())

	if err := s.Start(":8082"); err != nil {
		t.Fatal(err)
	}
}

func TestFileUploader_Handle(t *testing.T) {
	s := NewHttpServer()
	s.Get("/upload_page", func(ctx *Context) {
		tpl := template.New("upload")
		tpl, err := tpl.Parse(`
<html>
<body>
	<form action="/upload" method="post" enctype="multipart/form-data">
		 <input type="file" name="myfile" />
		 <button type="submit">上传</button>
	</form>
</body>
<html>
`)
		if err != nil {
			t.Fatal(err)
		}

		page := &bytes.Buffer{}
		err = tpl.Execute(page, nil)
		if err != nil {
			t.Fatal(err)
		}

		ctx.RespStatusCode = 200
		ctx.RespData = page.Bytes()
	})
	s.Post("/upload", (&FileUpload{
		// 这里的 myfile 就是 <input type="file" name="myfile" />
		// 那个 name 的取值
		FileField: "myfile",
		DstPathFunc: func(fh *multipart.FileHeader) string {
			return path.Join("testdata", "upload", fh.Filename)
		},
	}).Handler())
	s.Start(":8081")
}

func TestStaticResourceHandler_Handle(t *testing.T) {
	s := NewHttpServer()
	handler := NewStaticResourceHandler("./testdata/img", "/img")
	s.Get("/img/:file", handler.Handler)
	// 在浏览器里面输入 localhost:8081/img/come_on_baby.jpg
	s.Start(":8081")
}
