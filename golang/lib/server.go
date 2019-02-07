package lib

import (
	"fmt"
	"net"
	"sync"
)

type Server interface {
	Open(addr string) (net.Listener, error)
	Close() error
	WaitGroup() *sync.WaitGroup
	Listener() net.Listener
}

type BaseServer struct {
	Name      string
	WaitGroup *sync.WaitGroup
	Listener  net.Listener
}

func NewBaseServer(name string) *BaseServer {
	s := &BaseServer{
		Name:      name,
		WaitGroup: &sync.WaitGroup{},
	}
	return s
}

func (s *BaseServer) Open(addr string, callback func(net.Listener)) (net.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s.Listener = l
	s.WaitGroup.Add(1)
	go s.serve(l, callback)
	return l, nil
}

func (s *BaseServer) Close() error {
	if s.Listener == nil {
		return fmt.Errorf("%s listener isn't running", s.Name)
	}
	s.Listener.Close()
	s.WaitGroup.Wait()
	s.Listener = nil
	return nil
}

func (s *BaseServer) serve(l net.Listener, callback func(net.Listener)) {
	defer s.close()
	callback(l)
}

func (s *BaseServer) close() {
	s.Listener.Close()
	s.WaitGroup.Done()
}
