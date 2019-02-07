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

	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
)

const (
	DefaultHeartbeatTime = time.Minute
)

type ServerConfig struct {
	CacheAddr string
	WebAddr   string
	WebHost   string
}

type Server struct {
	base           *lib.BaseServer
	config         *ServerConfig
	client         *http.Client
	cache          *http.Client
	cacheAvailable bool
}

func New(config *ServerConfig) lib.Server {
	cacheURL, err := url.Parse("http://" + config.CacheAddr)
	if err != nil {
		panic(err)
	}
	cacheClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(cacheURL),
		},
	}
	return &Server{
		base:           lib.NewBaseServer("Controller"),
		config:         config,
		client:         http.DefaultClient,
		cache:          cacheClient,
		cacheAvailable: false,
	}
}

func (s *Server) Open(addr string) (net.Listener, error) {
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
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.RequestURI)
	u, err := url.Parse(req.RequestURI)
	if err != nil {
		httpHelpers.WriteError(w, 400, "Bad request URI")
		return
	}

	c := s.client
	host := req.Host
	hostname := host
	tokens := strings.SplitN(host, ":", 2)
	if len(tokens) >= 2 {
		hostname = tokens[0]
	}
	if host == s.config.WebHost {
		u.Host = s.config.WebAddr
	} else if strings.HasSuffix(hostname, ".granbluefantasy.jp") {
		u.Host = host
		// Hostname starting with 'game-a' usually meant for loading asset files
		if s.cacheAvailable && strings.HasPrefix(hostname, "game-a") {
			c = s.cache
		}
	} else {
		httpHelpers.WriteError(w, 403, "Host not allowed")
		return
	}

	if u.Scheme == "" {
		u.Scheme = "http"
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

func (s *Server) serve(l net.Listener) {
	go s.startCacheHeartbeat()
	err := http.Serve(l, s)
	if err != nil {
		// do nothing
	}
}

func (s *Server) startCacheHeartbeat() {
	req := &http.Request{
		Method: "GET",
		URL: &url.URL{
			Scheme: "http",
			Host:   s.config.CacheAddr,
		},
	}
	for s.Running() {
		s.cacheAvailable = s.checkCacheHeartbeat(req)
		time.Sleep(DefaultHeartbeatTime)
	}
}

func (s *Server) checkCacheHeartbeat(req *http.Request) bool {
	res, err := s.client.Do(req)
	if err != nil {
		log.Printf("Cache Heartbeat: Got error '%s'", err)
		return false
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Cache Heartbeat: Got error while reading response '%s'", err)
		return false
	}
	text := string(b)
	if text != "OK" {
		log.Printf("Cache Heartbeat: Expecting response 'OK', got '%s'", text)
		return false
	}
	log.Printf("Cache Heartbeat: %s", text)
	return true
}
