package main

import (
	"fmt"
	"testing"

	"unittest/gintest"
)

func Test_hello(t *testing.T) {
	objTest := gintest.Init()
	objTest.RegisterGet(hello, func(m *gintest.Mock) error {
		byteInfo := m.Exec()
		fmt.Printf("byteInfo--------------->"+"%+v\n", string(byteInfo))
		return nil
	})
}
