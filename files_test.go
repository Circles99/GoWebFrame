package GoWebFrame

import (
	"testing"
)

func TestFileDownload_Handler(t *testing.T) {

	s := NewHttpServer()
	s.Get("/download", (&FileDownload{Dir: "./testdata/download"}).Handler())

	if err := s.Start(":8082"); err != nil {
		t.Fatal(err)
	}
}
