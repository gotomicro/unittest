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
	storeRoute    map[string]routerPath
	storeCall     []string // 排序
	router        *gin.Engine
	middleware    []gin.HandlerFunc
	header        map[string]string
	tmpMiddleware []gin.HandlerFunc // 临时存储，最终放到router path
	tmpPath       string            // 临时存储，最终放到router path
}

type routerPath struct {
	Method     string
	f          func(c *gin.Context)
	call       func(m *Mock) error
	middleware []gin.HandlerFunc
	path       string
}

func Init(options ...TestOption) *Test {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	obj := &Test{
		storeRoute: make(map[string]routerPath),
		storeCall:  make([]string, 0),
		router:     router,
		header: map[string]string{
			"Content-Type": "application/json",
		},
	}
	for _, option := range options {
		option(obj)
	}
	return obj
}

func (t *Test) GET(f func(c *gin.Context), call func(m *Mock) error, options ...RouteOption) {
	t.register("GET", f, call, options...)
}

func (t *Test) POST(f func(c *gin.Context), call func(m *Mock) error, options ...RouteOption) {
	t.register("POST", f, call, options...)

}

func (t *Test) PUT(f func(c *gin.Context), call func(m *Mock) error, options ...RouteOption) {
	t.register("PUT", f, call, options...)

}

func (t *Test) DELETE(f func(c *gin.Context), call func(m *Mock) error, options ...RouteOption) {
	t.register("DELETE", f, call, options...)
}

func (t *Test) register(method string, f func(c *gin.Context), call func(m *Mock) error, options ...RouteOption) {
	path := urlPath()
	for _, option := range options {
		option(t)
	}
	if t.tmpPath != "" {
		path = t.tmpPath
		// 清空数据
		t.tmpPath = ""
	}
	var middleware []gin.HandlerFunc
	if len(t.tmpMiddleware) != 0 {
		middleware = t.tmpMiddleware
		// 清空数据
		t.tmpMiddleware = []gin.HandlerFunc{}
	}
	middleware = append(middleware, f)
	t.storeRoute[path] = routerPath{
		Method:     method,
		f:          f,
		call:       call,
		path:       path,
		middleware: middleware,
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
			t.router.GET(key, value.middleware...)
		case "POST":
			t.router.POST(key, value.middleware...)
		case "PUT":
			t.router.PUT(key, value.middleware...)
		case "DELETE":
			t.router.DELETE(key, value.middleware...)
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

func (m *Mock) Exec(options ...MockOption) []byte {
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

func WithQuery(data interface{}) MockOption {
	return func(c *Mock) {
		info, err := urlquery.Marshal(data)
		if err != nil {
			panic("with query failed,err: " + err.Error())
		}
		c.query = string(info)
	}
}

func WithJsonBody(data interface{}) MockOption {
	return func(c *Mock) {
		info, err := json.Marshal(data)
		if err != nil {
			panic("with json body failed,err: " + err.Error())
		}
		c.jsonBody = info
	}
}

func WithUri(uri string) MockOption {
	return func(c *Mock) {
		c.uri = uri
	}
}

func WithRoutePath(path string) RouteOption {
	return func(c *Test) {
		c.tmpPath = path
	}
}

func WithRouteMiddleware(middleware ...gin.HandlerFunc) RouteOption {
	return func(c *Test) {
		c.tmpMiddleware = middleware
	}
}

func WithTestMiddleware(middleware ...gin.HandlerFunc) TestOption {
	return func(c *Test) {
		c.middleware = middleware
	}
}

// MockOption 可选项
type MockOption func(c *Mock)
type RouteOption func(c *Test)
type TestOption func(c *Test)
