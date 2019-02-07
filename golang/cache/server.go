package cache

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
	"github.com/go-redis/redis"
	"github.com/vmihailenco/msgpack"
)

const DefaultExpirationTime = time.Hour

type ServerConfig struct {
	Redis *redis.Client
}

type Server struct {
	base   *lib.BaseServer
	config *ServerConfig
	redis  *redis.Client
}

type Cache struct {
	StatusCode    int
	Status        string
	Header        http.Header
	Body          []byte
	ContentLength int64
}

type CacheReader struct {
	Reader io.Reader
}

func New(config *ServerConfig) lib.Server {
	return &Server{
		base:   lib.NewBaseServer("Cache"),
		config: config,
		redis:  config.Redis,
	}
}

func (s *Server) Open(addr string) (net.Listener, error) {
	return s.base.Open(addr, s.serve)
}

func (s *Server) Close() error {
	return s.base.Close()
}

func (s *Server) Listener() net.Listener {
	return s.base.Listener
}

func (s *Server) WaitGroup() *sync.WaitGroup {
	return s.base.WaitGroup
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			httpHelpers.WriteServerError(w, 503, "Internal server error", err)
		}
	}()
	s.ServeHTTPUnsafe(w, req)
}

func (s *Server) ServeHTTPUnsafe(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	res, err := s.Fetch(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	for key, values := range res.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(res.StatusCode)
	_, err = w.Write(body)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
	}
	if ShouldCache(req, res) {
		s.CacheAsync(req, res, body, logError)
	}
}

func (s *Server) Fetch(req *http.Request) (*http.Response, error) {
	if s.HasCache(req) {
		return s.FetchFromCache(req)
	}
	return s.FetchFromServer(req)
}

func (s *Server) HasCache(req *http.Request) bool {
	key := GetKeyForRequest(req)
	val, err := s.redis.Exists(key).Result()
	if err != nil {
		panic(err)
	}
	return val == 1
}

func (s *Server) FetchFromCache(req *http.Request) (*http.Response, error) {
	key := GetKeyForRequest(req)
	b, err := s.redis.Get(key).Bytes()
	if err != nil {
		return nil, err
	}
	var cache *Cache
	err = msgpack.Unmarshal(b, &cache)
	if err != nil {
		return nil, err
	}
	body := &CacheReader{
		Reader: bytes.NewReader(cache.Body),
	}
	return &http.Response{
		StatusCode:    cache.StatusCode,
		Status:        cache.Status,
		Header:        cache.Header,
		Body:          body,
		ContentLength: cache.ContentLength,
	}, nil
}

func (s *Server) FetchFromServer(req *http.Request) (*http.Response, error) {
	u := req.URL
	if u.Host == "" {
		u.Host = req.Header.Get("Host")
	}
	c := http.DefaultClient
	return c.Do(&http.Request{
		URL:    u,
		Method: req.Method,
		Header: req.Header,
		Body:   req.Body,
	})
}

func (s *Server) CacheAsync(req *http.Request, res *http.Response, body []byte, callback func(error)) {
	go func() {
		callback(s.Cache(req, res, body))
	}()
}

func (s *Server) Cache(req *http.Request, res *http.Response, body []byte) error {
	cache, err := msgpack.Marshal(Cache{
		StatusCode:    res.StatusCode,
		Status:        res.Status,
		Header:        res.Header,
		Body:          body,
		ContentLength: res.ContentLength,
	})
	if err != nil {
		return err
	}
	key := GetKeyForRequest(req)
	status := s.redis.Set(key, cache, DefaultExpirationTime)
	return status.Err()
}

func (s *Server) serve(l net.Listener) {
	err := http.Serve(l, s)
	if err != nil {
		// do nothing
	}
}

func (c *CacheReader) Read(p []byte) (int, error) {
	return c.Reader.Read(p)
}

func (c *CacheReader) Close() error {
	return nil
}

func ShouldCache(req *http.Request, res *http.Response) bool {
	// TODO: Add logic on which request or response we should cache
	return true
}

func GetKeyForRequest(req *http.Request) string {
	return req.URL.Path
}

func logError(err error) {
	if err != nil {
		log.Println(err)
	}
}
