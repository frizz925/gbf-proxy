package local

import (
	"net"
	"net/http"
	"sync"

	"github.com/Frizz925/gbf-proxy/golang/proxy"

	"github.com/Frizz925/gbf-proxy/golang/cache"
	"github.com/Frizz925/gbf-proxy/golang/controller"
	"github.com/Frizz925/gbf-proxy/golang/lib"
)

type Server struct {
	Client     *http.Client
	Cache      lib.Server
	Controller lib.Server
	Proxy      lib.Server

	waitGroup *sync.WaitGroup
	listener  net.Listener
	servers   []lib.Server
	lock      *sync.Mutex
}

type ServerConfig struct {
	HttpClient *http.Client
}

func New(config *ServerConfig) *Server {
	return &Server{
		Client:    config.HttpClient,
		waitGroup: &sync.WaitGroup{},
		lock:      &sync.Mutex{},
	}
}

func (s *Server) Name() string {
	return "local"
}

func (s *Server) Open(addr string) (net.Listener, error) {
	defer s.lock.Unlock()
	s.lock.Lock()
	s.Cache = cache.New(&cache.ServerConfig{
		HttpClient: s.Client,
	})
	l, err := s.Cache.Open("localhost:0")
	if err != nil {
		return nil, err
	}
	s.servers = append(s.servers, s.Cache)
	s.Controller = controller.New(&controller.ServerConfig{
		CacheAddr:     l.Addr().String(),
		DefaultClient: s.Client,
	})
	l, err = s.Controller.Open("localhost:0")
	if err != nil {
		return nil, err
	}
	s.servers = append(s.servers, s.Controller)
	s.Proxy = proxy.New(&proxy.ServerConfig{
		BackendAddr: l.Addr().String(),
	})
	l, err = s.Proxy.Open(addr)
	if err != nil {
		return nil, err
	}
	s.servers = append(s.servers, s.Proxy)
	s.waitGroup.Add(1)

	go func() {
		for _, server := range s.servers {
			server.WaitGroup().Wait()
		}
		s.waitGroup.Done()
	}()
	return l, nil
}

func (s *Server) Close() error {
	for _, server := range s.servers {
		err := server.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) WaitGroup() *sync.WaitGroup {
	return s.waitGroup
}

func (s *Server) Listener() net.Listener {
	return s.listener
}

func (s *Server) Running() bool {
	for _, server := range s.GetServers() {
		if !server.Running() {
			return false
		}
	}
	return true
}

func (s *Server) GetServers() []lib.Server {
	defer s.lock.Unlock()
	s.lock.Lock()
	return s.servers
}
