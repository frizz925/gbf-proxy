package services

import (
	"gbf-proxy/lib/logger"
	"gbf-proxy/services/handlers"
	"io"
	"net"
)

type ProxyService struct {
	handlers.ConnectionForwarder
	log logger.Logger
}

var _ ListeningService = (*ProxyService)(nil)

func NewProxyService(c handlers.ConnectionForwarder) *ProxyService {
	return &ProxyService{
		ConnectionForwarder: c,
		log:                 logger.Factory.New(),
	}
}

func (s *ProxyService) Listen(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.HandleConnection(conn)
	}
}

func (s *ProxyService) HandleConnection(conn net.Conn) {
	defer conn.Close()
	err := s.ConnectionForwarder.ForwardConnection(conn)
	if err != nil && err != io.EOF {
		s.log.Error(err)
	}
}
