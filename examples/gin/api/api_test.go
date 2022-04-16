package api

import (
	"testing"

	"github.com/gotomicro/unittest/gintest"
	"github.com/stretchr/testify/assert"
)

func TestHello(t *testing.T) {
	objTest := gintest.Init()
	objTest.GET(Hello, func(m *gintest.Mock) error {
		byteInfo := m.Exec(gintest.WithQuery(HelloRequest{
			Name: "hello",
		}))
		assert.Equal(t, `Hello client: hello`, string(byteInfo))
		return nil
	})
	objTest.Run()
}
