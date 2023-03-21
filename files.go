package GoWebFrame

import (
	"fmt"
	"github.com/hashicorp/golang-lru"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileUpload 文件上传
type FileUpload struct {
	FileField string

	// 目标路径
	DstPathFunc func(fh *multipart.FileHeader) string
}

func (f FileUpload) Handler() HandleFunc {
	return func(c *Context) {
		src, srcHeader, err := c.Req.FormFile(f.FileField)
		if err != nil {
			c.RespStatusCode = 400
			c.RespData = []byte("上传失败，未找到数据")
			log.Fatalln(err)
			return
		}
		defer src.Close()

		dst, err := os.OpenFile(f.DstPathFunc(srcHeader), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 066)
		if err != nil {
			c.RespStatusCode = 500
			c.RespData = []byte("上传失败")
			log.Fatalln(err)
			return
		}

		defer dst.Close()
		// 把 src内容 copy到 dst中
		_, err = io.CopyBuffer(dst, src, nil)
		if err != nil {
			c.RespStatusCode = 500
			c.RespData = []byte("上传失败")
			log.Fatalln(err)
			return
		}
		c.RespData = []byte("上传成功")
	}
}

// FileDownload 文件下载
type FileDownload struct {
	Dir string
}

func (f FileDownload) Handler() HandleFunc {
	return func(c *Context) {
		req, _ := c.QueryValues("file").String()
		path := filepath.Join(f.Dir, filepath.Clean(req))

		// 获取后缀
		fn := filepath.Base(path)

		header := c.Resp.Header()
		header.Set("Content-Disposition", "attachment;filename="+fn)
		header.Set("Content-Description", "File Transfer")
		header.Set("Content-Type", "application/octet-stream")
		header.Set("Content-Transfer-Encoding", "binary")
		header.Set("Expires", "0")
		header.Set("Cache-Control", "must-revalidate")
		header.Set("Pragma", "public")

		// 获取文件
		http.ServeFile(c.Resp, c.Req, path)
	}
}

// StaticResourceHandler 静态文件
type StaticResourceHandler struct {
	dir                     string
	extensionContentTypeMap map[string]string

	//// 缓存静态资源的限制
	cache       *lru.Cache
	maxFileSize int
}

type fileCacheItem struct {
	fileName    string
	fileSize    int
	contentType string
	data        []byte
}

type StaticResourceHandlerOption func(h *StaticResourceHandler)

func NewStaticResourceHandler(dir string, pathPrefix string,
	options ...StaticResourceHandlerOption) *StaticResourceHandler {
	res := &StaticResourceHandler{
		dir: dir,
		extensionContentTypeMap: map[string]string{
			// 这里根据自己的需要不断添加
			"jpeg": "image/jpeg",
			"jpe":  "image/jpeg",
			"jpg":  "image/jpeg",
			"png":  "image/png",
			"pdf":  "image/pdf",
		},
	}

	for _, o := range options {
		o(res)
	}
	return res
}

func (s *StaticResourceHandler) Handler(c *Context) {

	req := c.PathParams["file"]

	if item, ok := s.readFileFromData(req); ok {
		s.writeItemAsResponse(item, c.Resp)
		return
	}

	path := filepath.Join(s.dir, req)

	f, err := os.Open(path)
	if err != nil {
		c.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	ext := getFileExt(f.Name())
	t, ok := s.extensionContentTypeMap[ext]
	if !ok {
		c.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := ioutil.ReadAll(f)

	if err != nil {
		c.Resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	item := &fileCacheItem{
		fileSize:    len(data),
		data:        data,
		contentType: t,
		fileName:    req,
	}
	s.cacheFile(item)
	s.writeItemAsResponse(item, c.Resp)
	return

}

func (s *StaticResourceHandler) writeItemAsResponse(item *fileCacheItem, writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", item.contentType)
	writer.Header().Set("Content-Length", fmt.Sprintf("%d", item.fileSize))
	_, _ = writer.Write(item.data)

}

func (s *StaticResourceHandler) cacheFile(item *fileCacheItem) {
	if s.cache != nil && item.fileSize < s.maxFileSize {
		s.cache.Add(item.fileName, item)
	}
}

func (s *StaticResourceHandler) readFileFromData(fileName string) (*fileCacheItem, bool) {
	if s.cache != nil {
		if item, ok := s.cache.Get(fileName); ok {
			return item.(*fileCacheItem), true
		}
	}
	return nil, false
}

func getFileExt(name string) string {
	index := strings.LastIndex(name, ".")
	if index == len(name)-1 {
		return ""
	}
	return name[index+1:]
}
