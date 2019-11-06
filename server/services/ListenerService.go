package services

import (
	connlib "gbf-proxy/lib/conn"
	"gbf-proxy/lib/logger"
	"gbf-proxy/services/handlers"
	"io"
	"net"
)

var log = logger.DefaultLogger

type ListenerService struct {
	Name string
	handlers.ConnectionForwarder
}

func NewListenerService(name string, c handlers.ConnectionForwarder) *ListenerService {
	return &ListenerService{
		Name:                name,
		ConnectionForwarder: c,
	}
}

func (s *ListenerService) Serve(addr string) error {
	l, err := connlib.CreateListener(addr)
	if err != nil {
		return err
	}
	defer l.Close()
	log.Infof("%s listening at %s", s.Name, addr)
	return s.Listen(l)
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
