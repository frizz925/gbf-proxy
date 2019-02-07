package controller

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Frizz925/gbf-proxy/golang/cache"
	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
)

const (
	DefaultHeartbeatTime = time.Minute
)

type ServerConfig struct {
	CacheAddr   string
	CacheClient *http.Client
	WebAddr     string
	WebHost     string
}

type Server struct {
	base           *lib.BaseServer
	config         *ServerConfig
	client         *http.Client
	cache          *http.Client
	cacheAvailable bool
	lock           *sync.Mutex
}

func New(config *ServerConfig) lib.Server {
	cacheClient := config.CacheClient
	if cacheClient == nil {
		cacheAddr := config.CacheAddr
		if cacheAddr == "" {
			log.Println("Cache address not set. Caching capability disabled.")
		} else {
			cacheClient = NewProxyClient(config.CacheAddr)
		}
	}
	webAddr := config.WebAddr
	if webAddr == "" {
		log.Println("Web address not set. Static web capability disabled.")
	}

	return &Server{
		base:           lib.NewBaseServer("Controller"),
		config:         config,
		client:         http.DefaultClient,
		cache:          cacheClient,
		cacheAvailable: cacheClient != nil,
		lock:           &sync.Mutex{},
	}
}

func (s *Server) Open(addr string) (net.Listener, error) {
	if s.CacheAvailable() {
		log.Printf("Controller at %s -> Cache service at %s", addr, s.config.CacheAddr)
	}
	if s.WebAvailable() {
		if s.config.WebHost == "" {
			log.Printf("Web hostname not set. Using the default %s.", addr)
			s.config.WebHost = addr
		}
		log.Printf("Controller at %s -> Web server at %s", addr, s.config.WebAddr)
	}
	return s.base.Open(addr, s.serve)
}

func (s *Server) Close() error {
	return s.base.Close()
}

func (s *Server) WaitGroup() *sync.WaitGroup {
	return s.base.WaitGroup
}

func (s *Server) Listener() net.Listener {
	return s.base.Listener
}

func (s *Server) Running() bool {
	return s.base.Running()
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			httpHelpers.WriteServerError(w, 503, "Internal server error", err)
		}
		req.Body.Close()
	}()
	s.ServeHTTPUnsafe(w, req)
}

func (s *Server) ServeHTTPUnsafe(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	u := httpHelpers.ParseURL(req)
	hostname := u.Host
	tokens := strings.SplitN(hostname, ":", 2)
	if len(tokens) >= 2 {
		hostname = tokens[0]
	}

	c := s.client
	if s.WebAvailable() && u.Host == s.config.WebHost {
		httpHelpers.LogRequest(s.base.Name, req, "Static web access")
		u.Host = s.config.WebAddr
	} else if strings.HasSuffix(hostname, ".granbluefantasy.jp") {
		// Hostname starting with 'game-a' usually meant for loading asset files
		if s.CacheAvailable() && strings.HasPrefix(hostname, "game-a") {
			c = s.cache
			httpHelpers.LogRequest(s.base.Name, req, "Cache access")
		} else {
			httpHelpers.LogRequest(s.base.Name, req, "Proxy access")
		}
	} else {
		httpHelpers.LogRequest(s.base.Name, req, "Forbidden host")
		httpHelpers.WriteError(w, 403, "Host not allowed")
		return
	}

	res, err := c.Do(&http.Request{
		Method: req.Method,
		URL:    u,
		Body:   req.Body,
		Header: req.Header,
	})
	if err != nil {
		httpHelpers.WriteServerError(w, 502, "Bad gateway", err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	for k, values := range res.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(res.StatusCode)
	length := len(body)
	for written := 0; written < length; {
		write, err := w.Write(body[written:])
		if err != nil {
			panic(err)
		}
		written += write
	}
}

func (s *Server) CacheAvailable() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.cache != nil && s.cacheAvailable
}

func (s *Server) WebAvailable() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.config.WebAddr != ""
}

func (s *Server) serve(l net.Listener) {
	go s.startCacheHeartbeat()
	err := http.Serve(l, s)
	if err != nil {
		// do nothing
	}
}

func (s *Server) startCacheHeartbeat() {
	header := make(http.Header)
	header.Set(cache.CacheAPIHeaderName, "1")
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host:   s.config.CacheAddr,
		},
		Header: header,
	}
	for s.Running() {
		cacheAvailable := false
		if s.cache != nil {
			cacheAvailable = s.checkCacheHeartbeat(req)
		}
		s.lock.Lock()
		s.cacheAvailable = cacheAvailable
		s.lock.Unlock()
		time.Sleep(DefaultHeartbeatTime)
	}
}

func (s *Server) checkCacheHeartbeat(req *http.Request) bool {
	res, err := s.cache.Do(req)
	if err != nil {
		log.Printf("Cache Heartbeat: Got error '%s'", err)
		return false
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Cache Heartbeat: Got error while reading response '%s'", err)
		return false
	}
	text := strings.TrimSpace(string(b))
	if text != "OK" {
		log.Printf("Cache Heartbeat: Expecting response 'OK', got '%s'", text)
		return false
	}
	log.Printf("Cache Heartbeat: %s", text)
	return true
}

func NewProxyClient(host string) *http.Client {
	cacheURL := &url.URL{
		Scheme: "http",
		Host:   host,
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(cacheURL),
		},
	}
}
