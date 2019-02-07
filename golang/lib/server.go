package lib

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type Server interface {
	Open(addr string) (net.Listener, error)
	Close() error
	WaitGroup() *sync.WaitGroup
	Listener() net.Listener
	Running() bool
}

type BaseServer struct {
	Name      string
	WaitGroup *sync.WaitGroup
	Listener  net.Listener
	running   bool
	stateSync *sync.WaitGroup
}

func NewBaseServer(name string) *BaseServer {
	s := &BaseServer{
		Name:      name,
		WaitGroup: &sync.WaitGroup{},
		running:   false,
		stateSync: &sync.WaitGroup{},
	}
	return s
}

func (s *BaseServer) Open(addr string, callback func(net.Listener)) (net.Listener, error) {
	if s.Running() {
		return nil, fmt.Errorf("%s service already running", s.Name)
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s.running = true
	s.Listener = l
	s.WaitGroup.Add(1)
	go s.serve(l, callback)
	log.Printf("%s service listening at %s", s.Name, l.Addr().String())
	return l, nil
}

func (s *BaseServer) Close() error {
	if !s.Running() {
		return fmt.Errorf("%s listener isn't running", s.Name)
	}
	s.Listener.Close()
	s.WaitGroup.Wait()
	s.running = false
	s.Listener = nil
	return nil
}

func (s *BaseServer) Running() bool {
	s.stateSync.Wait()
	s.stateSync.Add(1)
	defer s.stateSync.Done()
	return s.running && s.Listener != nil
}

func (s *BaseServer) serve(l net.Listener, callback func(net.Listener)) {
	defer s.close()
	callback(l)
}

func (s *BaseServer) close() {
	s.Listener.Close()
	s.WaitGroup.Done()
}
