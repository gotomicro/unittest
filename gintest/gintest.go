package gintest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/askuy/urlquery"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Test struct {
	storeRoute    map[string]map[string]routerPath // map<METHOD, map<PATH, routerPath>>
	storeCall     []routerPath                     // 有序列表
	router        *gin.Engine                      // router
	middleware    []gin.HandlerFunc                // middleware
	header        map[string]string                // 请求header
	tmpMiddleware []gin.HandlerFunc                // 临时存储，最终放到router path
	tmpPath       string                           // 临时存储，最终放到router path
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
		storeRoute: make(map[string]map[string]routerPath),
		storeCall:  make([]routerPath, 0),
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

func (t *Test) PATCH(f func(c *gin.Context), call func(m *Mock) error, options ...RouteOption) {
	t.register("PATCH", f, call, options...)
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
	_, ok := t.storeRoute[method]
	if !ok {
		t.storeRoute[method] = make(map[string]routerPath)
	}
	rp := routerPath{
		Method:     method,
		f:          f,
		call:       call,
		path:       path,
		middleware: middleware,
	}
	t.storeRoute[method][path] = rp
	t.storeCall = append(t.storeCall, rp)
}

func urlPath() string {
	return strings.ReplaceAll("/"+uuid.New().String(), "-", "")
}

func (t *Test) Run() error {
	t.router.Use(t.middleware...)
	for _, methodRoutePaths := range t.storeRoute {
		for path, value := range methodRoutePaths {
			switch value.Method {
			case "GET":
				t.router.GET(path, value.middleware...)
			case "POST":
				t.router.POST(path, value.middleware...)
			case "PUT":
				t.router.PUT(path, value.middleware...)
			case "PATCH":
				t.router.PATCH(path, value.middleware...)
			case "DELETE":
				t.router.DELETE(path, value.middleware...)
			}
		}
	}

	for _, key := range t.storeCall {
		methodRouts, ok := t.storeRoute[key.Method]
		if !ok {
			panic("method not exist")
		}
		storePath, flag := methodRouts[key.path]
		if !flag {
			panic("path not exist")
		}

		mock := &Mock{
			path:   key.path,
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
	path   string
	method string
	router *gin.Engine
	header map[string]string
	query  string
	body   []byte
}

// Run 返回完整http.Response
func (m *Mock) Run(options ...MockOption) *http.Response {
	for _, option := range options {
		option(m)
	}
	path := m.path
	if m.query != "" {
		path = path + "?" + m.query
	}

	var req *http.Request
	if len(m.body) != 0 {
		reader := bytes.NewReader(m.body)
		req = httptest.NewRequest(m.method, path, reader)
	} else {
		req = httptest.NewRequest(m.method, path, nil)
	}

	for key, value := range m.header {
		req.Header.Set(key, value)
	}

	// 初始化响应
	w := CreateTestResponseRecorder()

	// 调用相应handler接口
	m.router.ServeHTTP(w, req)

	return w.Result()
}

// Exec 仅返回response body
func (m *Mock) Exec(options ...MockOption) []byte {
	result := m.Run(options...)
	defer result.Body.Close()

	// 读取响应body
	body, _ := io.ReadAll(result.Body)
	return body
}

// WithQuery 设置请求的query参数
func WithQuery(data interface{}) MockOption {
	return func(c *Mock) {
		info, err := urlquery.Marshal(data)
		if err != nil {
			panic("with query failed,err: " + err.Error())
		}
		c.query = string(info)
	}
}

// WithJsonBody 设置请求的body，会自动转成json字符串
func WithJsonBody(data interface{}) MockOption {
	return func(c *Mock) {
		info, err := json.Marshal(data)
		if err != nil {
			panic("with json body failed,err: " + err.Error())
		}
		c.body = info
	}
}

// WithBody 设置请求的body
func WithBody(data []byte) MockOption {
	return func(c *Mock) {
		c.body = data
	}
}

// WithUri 设置请求的uri
func WithUri(uri string) MockOption {
	return func(c *Mock) {
		c.path = uri
	}
}

// WithHeader 设置请求的header
func WithHeader(key, value string) MockOption {
	return func(c *Mock) {
		c.header[key] = value
	}
}

// WithRoutePath 设置请求的注册path
func WithRoutePath(path string) RouteOption {
	return func(c *Test) {
		c.tmpPath = path
	}
}

// WithRouteMiddleware 设置请求注册的中间件
func WithRouteMiddleware(middleware ...gin.HandlerFunc) RouteOption {
	return func(c *Test) {
		c.tmpMiddleware = middleware
	}
}

// WithTestMiddleware 开始调试时，用例注入的中间件
func WithTestMiddleware(middleware ...gin.HandlerFunc) TestOption {
	return func(c *Test) {
		c.middleware = middleware
	}
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func (r *TestResponseRecorder) closeClient() {
	r.closeChannel <- true
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

// MockOption 可选项
type MockOption func(c *Mock)
type RouteOption func(c *Test)
type TestOption func(c *Test)

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}
