package applications

import (
	"fmt"
)

type HelloWorldApp struct {
}

var _ Application = (*HelloWorldApp)(nil)

func NewHelloWorldApp() *HelloWorldApp {
	return &HelloWorldApp{}
}

func (HelloWorldApp) Start() {
	fmt.Println("Hello, world!")
}
