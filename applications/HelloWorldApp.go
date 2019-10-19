package applications

import (
	"log"
	"net"

	"github.com/Frizz925/gbf-proxy/services"
)

type HelloWorldApp struct {
}

var _ Application = (*HelloWorldApp)(nil)

func NewHelloWorldApp() *HelloWorldApp {
	return &HelloWorldApp{}
}

func (HelloWorldApp) Start() {
	service := services.NewHelloWorldService()
	listener, err := net.Listen("tcp4", "127.0.0.1:8000")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Service listening at :8000")
	service.Listen(listener)
}
