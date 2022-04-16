package gintest

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/askuy/urlquery"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Test struct {
	storeRoute map[string]routerPath
	storeCall  []string // 排序
	router     *gin.Engine
	middleware []gin.HandlerFunc
	header     map[string]string
}

type routerPath struct {
	Method string
	f      func(c *gin.Context)
	call   func(m *Mock) error
}

func Init(middleware ...gin.HandlerFunc) *Test {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	return &Test{
		storeRoute: make(map[string]routerPath),
		storeCall:  make([]string, 0),
		router:     router,
		middleware: middleware,
		header: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

func (t *Test) GET(f func(c *gin.Context), call func(m *Mock) error) {
	t.register("GET", f, call)

}

func (t *Test) POST(f func(c *gin.Context), call func(m *Mock) error) {
	t.register("POST", f, call)

}

func (t *Test) PUT(f func(c *gin.Context), call func(m *Mock) error) {
	t.register("PUT", f, call)

}

func (t *Test) DELETE(f func(c *gin.Context), call func(m *Mock) error) {
	t.register("DELETE", f, call)
}

func (t *Test) register(method string, f func(c *gin.Context), call func(m *Mock) error) {
	path := urlPath()
	t.storeRoute[path] = routerPath{
		Method: method,
		f:      f,
		call:   call,
	}
	t.storeCall = append(t.storeCall, path)
}

func urlPath() string {
	return strings.ReplaceAll("/"+uuid.New().String(), "-", "")
}

func (t *Test) Run() error {
	t.router.Use(t.middleware...)
	for key, value := range t.storeRoute {
		switch value.Method {
		case "GET":
			t.router.GET(key, value.f)
		case "POST":
			t.router.POST(key, value.f)
		case "PUT":
			t.router.PUT(key, value.f)
		case "DELETE":
			t.router.DELETE(key, value.f)
		}
	}

	for _, key := range t.storeCall {
		storePath, flag := t.storeRoute[key]
		if !flag {
			panic("url not exist")
		}

		mock := &Mock{
			uri:    key,
			method: storePath.Method,
			router: t.router,
			header: t.header,
		}
		err := storePath.call(mock)
		if err != nil {
			return err
		}
	}
	return nil
}

type Mock struct {
	uri      string
	method   string
	router   *gin.Engine
	header   map[string]string
	query    string
	jsonBody []byte
}

func (m *Mock) Exec(options ...Option) []byte {
	for _, option := range options {
		option(m)
	}
	uri := m.uri
	if m.query != "" {
		uri = uri + "?" + m.query
	}
	var req *http.Request
	if len(m.jsonBody) != 0 {
		reader := bytes.NewReader(m.jsonBody)
		req = httptest.NewRequest(m.method, uri, reader)
	} else {
		req = httptest.NewRequest(m.method, uri, nil)
	}

	for key, value := range m.header {
		req.Header.Set(key, value)
	}

	// 初始化响应
	w := httptest.NewRecorder()

	// 调用相应handler接口
	m.router.ServeHTTP(w, req)

	// 提取响应
	result := w.Result()
	defer result.Body.Close()

	// 读取响应body
	body, _ := ioutil.ReadAll(result.Body)
	return body
}

func WithQuery(data interface{}) Option {
	return func(c *Mock) {
		info, err := urlquery.Marshal(data)
		if err != nil {
			panic("with query failed,err: " + err.Error())
		}
		c.query = string(info)
	}
}

func WithJsonBody(data interface{}) Option {
	return func(c *Mock) {
		info, err := json.Marshal(data)
		if err != nil {
			panic("with json body failed,err: " + err.Error())
		}
		c.jsonBody = info
	}
}

// Option 可选项
type Option func(c *Mock)
