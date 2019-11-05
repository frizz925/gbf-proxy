package services

import (
	"gbf-proxy/lib/logger"
	"gbf-proxy/services/handlers"
	"io"
	"net"
)

var log = logger.DefaultLogger

type ListenerService struct {
	handlers.ConnectionForwarder
}

func NewListenerService(c handlers.ConnectionForwarder) *ListenerService {
	return &ListenerService{
		ConnectionForwarder: c,
	}
}

func (s *ListenerService) Listen(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.HandleConnection(conn)
	}
}

func (s *ListenerService) HandleConnection(conn net.Conn) {
	defer conn.Close()
	err := s.ConnectionForwarder.ForwardConnection(conn)
	if err != nil && err != io.EOF {
		log.Error(err)
	}
}
