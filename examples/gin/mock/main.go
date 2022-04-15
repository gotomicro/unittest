package main

import (
	"fmt"

	"github.com/gotomicro/unittest/examples/gin/api"
	"github.com/gotomicro/unittest/gintest"
)

func main() {
	objTest := gintest.Init()
	objTest.RegisterGet(api.Hello, func(m *gintest.Mock) error {
		byteInfo := m.Exec(gintest.WithQuery(api.HelloRequest{
			Name: "hello",
		}))
		fmt.Printf("byteInfo--------------->"+"%+v\n", string(byteInfo))
		return nil
	})
	objTest.RegisterPost(api.PostHello, func(m *gintest.Mock) error {
		byteInfo := m.Exec(gintest.WithJsonBody(api.PostHelloRequest{
			Name: "hello",
		}))
		fmt.Printf("byteInfo--------------->"+"%+v\n", string(byteInfo))
		return nil
	})
	objTest.Run()
}
