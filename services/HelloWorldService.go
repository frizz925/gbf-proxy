package services

import (
	"log"
	"net"
)

type HelloWorldService struct {
	listener net.Listener
}

var _ ListeningService = (*HelloWorldService)(nil)

func NewHelloWorldService() *HelloWorldService {
	return &HelloWorldService{}
}

func (HelloWorldService) Listen(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		log.Panic(err)
	}
	conn.Write([]byte("Hello, world!"))
	err = conn.Close()
	if err != nil {
		log.Panic(err)
	}
	err = listener.Close()
	if err != nil {
		log.Panic(err)
	}
}
