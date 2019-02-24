package controller

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Frizz925/gbf-proxy/golang/lib/logging"

	"github.com/Frizz925/gbf-proxy/golang/cache"
	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
	"github.com/jinzhu/copier"
)

const (
	DefaultHeartbeatTime = time.Minute
)

type ServerConfig struct {
	CacheAddr     string
	DefaultClient *http.Client
	CacheClient   *http.Client
	WebAddr       string
	WebHost       string
}

type Server struct {
	base           *lib.BaseServer
	config         *ServerConfig
	client         *http.Client
	cache          *http.Client
	cacheAvailable bool
	lock           *sync.Mutex
	logger         *logging.Logger
}

func New(config *ServerConfig) lib.Server {
	logger := logging.New(&logging.LoggerConfig{
		Name: "Controller",
	})
	cacheClient := config.CacheClient
	if cacheClient == nil {
		cacheAddr := config.CacheAddr
		if cacheAddr == "" {
			logger.Info("Cache address not set. Caching capability disabled.")
		} else {
			cacheClient = NewProxyClient(config.CacheAddr)
		}
	}
	webAddr := config.WebAddr
	if webAddr == "" {
		logger.Info("Web address not set. Static web capability disabled.")
	}
	client := config.DefaultClient
	if client == nil {
		client = http.DefaultClient
	}

	return &Server{
		base:           lib.NewBaseServer(logger.Name),
		config:         config,
		client:         client,
		cache:          cacheClient,
		cacheAvailable: cacheClient != nil,
		lock:           &sync.Mutex{},
		logger:         logger,
	}
}

func (s *Server) Open(addr string) (net.Listener, error) {
	if s.CacheAvailable() {
		s.logger.Infof("Controller at %s -> Cache service at %s", addr, s.config.CacheAddr)
	}
	if s.WebAvailable() {
		if s.config.WebHost == "" {
			hostname := httpHelpers.AddrToHost(addr)
			s.logger.Infof("Web hostname not set. Using the default %s.", hostname)
			s.config.WebHost = hostname
		}
		s.logger.Infof("Controller at %s -> Web server at %s", addr, s.config.WebAddr)
	}
	return s.base.Open(addr, s.serve)
}

func (s *Server) Name() string {
	return s.base.Name
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
	hostname := u.Hostname()

	c := s.client
	if s.WebAvailable() && hostname == s.config.WebHost {
		httpHelpers.LogRequest(s.logger, req, "Static web access")
		u.Host = s.config.WebAddr
	} else if strings.HasSuffix(hostname, ".granbluefantasy.jp") {
		// Hostname starting with 'game-a' usually meant for loading asset files
		if s.CacheAvailable() && strings.HasPrefix(hostname, "game-a") {
			c = s.cache
			httpHelpers.LogRequest(s.logger, req, "Cache access")
		} else {
			httpHelpers.LogRequest(s.logger, req, "Proxy access")
		}
	} else {
		httpHelpers.LogRequest(s.logger, req, "Forbidden host")
		httpHelpers.WriteError(w, 403, "Host not allowed")
		return
	}

	clientReq := &http.Request{}
	err := copier.Copy(clientReq, req)
	if err != nil {
		panic(err)
	}
	clientReq.RequestURI = ""

	res, err := c.Do(clientReq)
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
	for sent := 0; sent < length; {
		written, err := w.Write(body[sent:])
		if err != nil {
			panic(err)
		}
		sent += written
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
			Path:   "/ping",
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
		s.logger.Infof("Cache Heartbeat: Got error '%s'", err)
		return false
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		s.logger.Infof("Cache Heartbeat: Got error while reading response '%s'", err)
		return false
	}
	text := strings.TrimSpace(string(b))
	if text != "OK" {
		s.logger.Infof("Cache Heartbeat: Expecting response 'OK', got '%s'", text)
		return false
	}
	s.logger.Infof("Cache Heartbeat: %s", text)
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
