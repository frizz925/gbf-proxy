package controller

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/vmihailenco/msgpack"

	"github.com/Frizz925/gbf-proxy/golang/cache"
	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
	wsHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/websocket"
	"github.com/Frizz925/gbf-proxy/golang/lib/websocket"
	"github.com/jinzhu/copier"
)

const (
	DefaultHeartbeatTime = time.Minute
	WritePeriod          = time.Second * 30
	PingPeriod           = time.Second * 60
	ReadBufferSize       = 4096
	WriteBufferSize      = 4096
)

type IncomingRequest = wsHelpers.Request
type OutgoingResponse = wsHelpers.Response

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
	wsLock         *sync.Mutex
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
		upgrader: &websocket.Upgrader{
			EnableCompression: true,
			ReadBufferSize:    ReadBufferSize,
			WriteBufferSize:   WriteBufferSize,
		},
		wsLock: &sync.Mutex{},
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
	err := s.ServeHTTPUnsafe(w, req)
	if err == nil {
		return
	}

	code := 503
	message := "Internal server error"
	if reqErr, ok := err.(*httpHelpers.RequestError); ok {
		code = reqErr.StatusCode
		message = reqErr.Message
	}
	httpHelpers.WriteServerError(s.base.Logger, w, code, message, err)
}

func (s *Server) ServeHTTPUnsafe(w http.ResponseWriter, req *http.Request) error {
	upgrade := req.Header.Get("Upgrade")
	if upgrade == "websocket" {
		ws, err := s.upgrader.Upgrade(w, req, nil)
		if err != nil {
			return err
		}
		httpHelpers.LogRequest(s.base.Logger, req, "Upgrading to WebSocket")
		s.ListenWebSocket(ws)
		return nil
	}
	defer req.Body.Close()

	res, err := s.ForwardRequest(req)
	if err != nil {
		if _, ok := err.(*httpHelpers.RequestError); ok {
			return err
		}
		return httpHelpers.NewRequestError(502, "Bad gateway", err)
	}
	defer res.Body.Close()

	for k, values := range res.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(res.StatusCode)

	buffer := make([]byte, 8192)
	for finished := false; !finished; {
		length, err := res.Body.Read(buffer)
		if err == io.EOF {
			finished = true
		} else if err != nil {
			return err
		}
		for sent := 0; sent < length; {
			written, err := w.Write(buffer[sent:length])
			if err == io.EOF {
				finished = true
				break
			} else if err != nil {
				return err
			}
			sent += written
		}
	}
	return nil
}

func (s *Server) ListenWebSocket(ws *websocket.Conn) {
	defer ws.Close()

	ctx := websocket.NewContext(ws)
	for s.Running() && ctx.Connected() {
		err := s.ServeWebSocket(ctx)
		if err != nil {
			s.base.Logger.Error(err)
			break
		}
	}
}

func (s *Server) ServeWebSocket(ctx *websocket.Context) error {
	// Receive the incoming request
	data, err := ctx.Read()
	if err != nil {
		return err
	}

	// Unmarshal and forward the request
	var r IncomingRequest
	err = msgpack.Unmarshal(data, &r)
	if err != nil {
		return err
	}
	req, err := httpHelpers.UnserializeRequest(&r.Payload)
	if err != nil {
		return err
	}

	req.RemoteAddr = ctx.Conn.RemoteAddr().String()
	go s.handleWebSocketRequest(ctx, r.ID, req)
	return nil
}

func (s *Server) handleWebSocketRequest(ctx *websocket.Context, id string, req *http.Request) {
	err := s.handleWebSocketRequestUnsafe(ctx, id, req)
	if err != nil {
		s.base.Logger.Error(err)
	}
}

func (s *Server) handleWebSocketRequestUnsafe(ctx *websocket.Context, id string, req *http.Request) error {
	res, err := s.ForwardRequest(req)
	if err != nil {
		return err
	}

	// Marshal and return the response
	r, err := httpHelpers.SerializeResponse(res)
	if err != nil {
		return err
	}
	data, err := msgpack.Marshal(OutgoingResponse{
		ID:      id,
		Payload: *r,
	})
	if err != nil {
		return err
	}
	return ctx.Write(data)
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
		return nil, httpHelpers.NewRequestError(403, "Host not allowed", nil)
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
