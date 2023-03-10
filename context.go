package GoWebFrame

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type Context struct {
	Req              *http.Request
	Resp             http.ResponseWriter
	PathParams       map[string]string
	cacheQueryValues url.Values
	RespStatusCode   int
	RespData         []byte
	MatchedRoute     string // 匹配的路由
}

type StringValue struct {
	val string
	err error
}

// BindJson
// @Description: 绑定json
// @receiver c
// @param val
// @return error
func (c *Context) BindJson(val any) error {
	if val == nil {
		return errors.New("web: 错误的绑定值")
	}
	decoder := json.NewDecoder(c.Req.Body)
	return decoder.Decode(val)
}

// FormValue
// @Description: 获取from中的值
// @receiver c
// @param key
// @return StringValue
func (c *Context) FormValue(key string) StringValue {
	// 解析from
	if err := c.Req.ParseForm(); err != nil {
		return StringValue{err: err}
	}
	return StringValue{val: c.Req.FormValue(key)}
}

// QueryValues
// @Description: 获取value值
// @receiver c
// @param key
// @return StringValue
func (c *Context) QueryValues(key string) StringValue {
	// 解析urlQuery
	// 但因为不可每次都调用Url.query, 所以缓存
	if c.cacheQueryValues == nil {
		c.cacheQueryValues = c.Req.URL.Query()
	}

	val, ok := c.cacheQueryValues[key]
	if !ok {
		return StringValue{err: errors.New("web: 找不到这个 key")}
	}
	return StringValue{val: val[0]}
}

// RespJson
// @Description: 返回json
// @receiver c
// @param val
// @return error
func (c *Context) RespJson(val any) error {
	_, err := json.Marshal(val)
	if err != nil {
		return err
	}
	c.Resp.Header().Set("Content-Type", "application/json")
	return nil
}
