package controller

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

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
	upgrader       *websocket.Upgrader
}

type RequestError struct {
	base       error
	StatusCode int
	Message    string
}

func NewRequestError(code int, message string, err error) *RequestError {
	return &RequestError{
		base:       err,
		StatusCode: code,
		Message:    message,
	}
}

func (e *RequestError) Error() string {
	if e.base != nil {
		return e.base.Error()
	}
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

func New(config *ServerConfig) lib.Server {
	base := lib.NewBaseServer("Controller")
	cacheClient := config.CacheClient
	if cacheClient == nil {
		cacheAddr := config.CacheAddr
		if cacheAddr == "" {
			base.Logger.Info("Cache address not set. Caching capability disabled.")
		} else {
			cacheClient = NewProxyClient(config.CacheAddr)
		}
	}
	webAddr := config.WebAddr
	if webAddr == "" {
		base.Logger.Info("Web address not set. Static web capability disabled.")
	}
	client := config.DefaultClient
	if client == nil {
		client = http.DefaultClient
	}

	return &Server{
		base:           base,
		config:         config,
		client:         client,
		cache:          cacheClient,
		cacheAvailable: cacheClient != nil,
		lock:           &sync.Mutex{},
		upgrader:       &websocket.Upgrader{},
	}
}

func (s *Server) Open(addr string) (net.Listener, error) {
	if s.CacheAvailable() {
		s.base.Logger.Infof("Controller service at %s -> Cache service at %s", addr, s.config.CacheAddr)
	}
	if s.WebAvailable() {
		if s.config.WebHost == "" {
			hostname := httpHelpers.AddrToHost(addr)
			s.base.Logger.Infof("Web hostname not set. Using the default %s.", hostname)
			s.config.WebHost = hostname
		}
		s.base.Logger.Infof("Controller service at %s -> Web server at %s", addr, s.config.WebAddr)
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
			code := 503
			message := "Internal server error"
			if reqErr, ok := r.(*RequestError); ok {
				code = reqErr.StatusCode
				message = reqErr.Message
			}
			httpHelpers.WriteServerError(s.base.Logger, w, code, message, err)
		}
		req.Body.Close()
	}()
	s.ServeHTTPUnsafe(w, req)
}

func (s *Server) ServeHTTPUnsafe(w http.ResponseWriter, req *http.Request) {
	upgrade := req.Header.Get("Upgrade")
	if upgrade == "websocket" {
		ws, err := s.upgrader.Upgrade(w, req, nil)
		if err != nil {
			panic(err)
		}
		s.ServeWebSocket(ws)
		return
	}
	defer req.Body.Close()

	res, err := s.ForwardRequest(req)
	if err != nil {
		if _, ok := err.(*RequestError); ok {
			panic(err)
		}
		httpHelpers.WriteServerError(s.base.Logger, w, 502, "Bad gateway", err)
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

func (s *Server) ServeWebSocket(ws *websocket.Conn) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		err := r.(error)
		s.base.Logger.Error(err)

		code := 503
		message := "Internal server error"
		if reqErr, ok := r.(*RequestError); ok {
			code = reqErr.StatusCode
			message = reqErr.Message
		}

		body := []byte(message)
		err = s.WriteToWebSocket(ws, &http.Response{
			StatusCode: code,
			Body:       httpHelpers.NewBodyReader(body),
		})
		if err != nil {
			s.base.Logger.Error(err)
		}
	}()
	s.ServeWebSocketUnsafe(ws)
}

func (s *Server) ServeWebSocketUnsafe(ws *websocket.Conn) {
	defer ws.Close()
	req, err := s.ReadFromWebSocket(ws)
	if err != nil {
		panic(err)
	}
	res, err := s.ForwardRequest(req)
	if err != nil {
		panic(err)
	}
	err = s.WriteToWebSocket(ws, res)
	if err != nil {
		panic(err)
	}
}

func (s *Server) ReadFromWebSocket(ws *websocket.Conn) (*http.Request, error) {
	_, data, err := ws.ReadMessage()
	if err != nil {
		return nil, err
	}
	return s.UnmarshalRequest(data)
}

func (s *Server) WriteToWebSocket(ws *websocket.Conn, res *http.Response) error {
	data, err := s.MarshalResponse(res)
	if err != nil {
		return err
	}
	return ws.WriteMessage(websocket.BinaryMessage, data)
}

func (s *Server) UnmarshalRequest(data []byte) (*http.Request, error) {
	return nil, errors.New("Not implemented yet")
}

func (s *Server) MarshalResponse(res *http.Response) ([]byte, error) {
	return nil, errors.New("Not implemented yet")
}

func (s *Server) ForwardRequest(req *http.Request) (*http.Response, error) {
	u := httpHelpers.ParseURL(req)
	hostname := u.Hostname()

	c := s.client
	if s.WebAvailable() && hostname == s.config.WebHost {
		httpHelpers.LogRequest(s.base.Logger, req, "Static web access")
		u.Host = s.config.WebAddr
	} else if strings.HasSuffix(hostname, ".granbluefantasy.jp") {
		// Hostname starting with 'game-a' usually meant for loading asset files
		if s.CacheAvailable() && strings.HasPrefix(hostname, "game-a") {
			c = s.cache
			httpHelpers.LogRequest(s.base.Logger, req, "Cache access")
		} else {
			httpHelpers.LogRequest(s.base.Logger, req, "Proxy access")
		}
	} else {
		httpHelpers.LogRequest(s.base.Logger, req, "Forbidden host")
		return nil, NewRequestError(403, "Host not allowed", nil)
	}

	clientReq := &http.Request{}
	err := copier.Copy(clientReq, req)
	if err != nil {
		panic(err)
	}
	clientReq.RequestURI = ""
	return c.Do(clientReq)
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
		s.base.Logger.Error(err)
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
		s.base.Logger.Infof("Cache Heartbeat: Got error '%s'", err)
		return false
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		s.base.Logger.Infof("Cache Heartbeat: Got error while reading response '%s'", err)
		return false
	}
	text := strings.TrimSpace(string(b))
	if text != "OK" {
		s.base.Logger.Infof("Cache Heartbeat: Expecting response 'OK', got '%s'", text)
		return false
	}
	s.base.Logger.Infof("Cache Heartbeat: %s", text)
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
