package proxy

import (
	"net"
	"net/http"
	"sync"
)

type Server struct {
	WaitGroup *sync.WaitGroup
	Listener  net.Listener
}

func NewServer() *Server {
	s := &Server{}
	s.WaitGroup = &sync.WaitGroup{}
	return s
}

func (s *Server) Start(addr string) (net.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s.Listener = l
	s.WaitGroup.Add(1)
	go s.serve()
	return l, nil
}

func (s *Server) Stop() {
	if s.Listener == nil {
		panic("Server isn't running")
	}
	s.Listener.Close()
	s.WaitGroup.Wait()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	_, err := w.Write([]byte("Welcome to Granblue Proxy 0.1-alpha!"))
	if err != nil {
		panic(err)
	}
}

func (s *Server) serve() {
	defer s.close()
	err := http.Serve(s.Listener, s)
	if err != nil {
		panic(err)
	}
}

func (s *Server) close() {
	s.Listener.Close()
	s.WaitGroup.Done()
}
