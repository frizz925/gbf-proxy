package cache

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/Frizz925/gbf-proxy/golang/lib"
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
	"github.com/go-redis/redis"
	"github.com/vmihailenco/msgpack"
)

const (
	DefaultExpirationTime = time.Hour
	DefaultHeartbeatTime  = time.Minute
	CleanUpIntervalTime   = time.Minute
	CacheAPIHeaderName    = "X-Granblue-Cache-API"
)

type ServerConfig struct {
	RedisAddr string
	Redis     *redis.Client
}

type Server struct {
	base           *lib.BaseServer
	config         *ServerConfig
	cache          *cache.Cache
	redis          *redis.Client
	redisAvailable bool
	lock           *sync.Mutex
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
	internalCache := cache.New(DefaultExpirationTime, CleanUpIntervalTime)
	redisClient := config.Redis
	if redisClient == nil {
		redisAddr := config.RedisAddr
		if redisAddr == "" {
			log.Printf("Redis address not set. Falling back to built-in in-memory caching.")
		} else {
			redisClient = redis.NewClient(&redis.Options{
				Addr:     redisAddr,
				Password: "",
				DB:       0,
			})
		}
	}

	return &Server{
		base:           lib.NewBaseServer("Cache"),
		config:         config,
		cache:          internalCache,
		redis:          redisClient,
		redisAvailable: redisClient != nil,
		lock:           &sync.Mutex{},
	}
}

func (s *Server) Open(addr string) (net.Listener, error) {
	if s.RedisAvailable() {
		log.Printf("Cache at %s -> Redis server at %s", addr, s.config.RedisAddr)
	}
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

func (s *Server) Running() bool {
	return s.base.Running()
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

	apiSwitch := req.Header.Get(CacheAPIHeaderName)
	if apiSwitch == "1" {
		s.ServeAsAPI(w, req)
		return
	}

	u := httpHelpers.ParseURL(req)
	if u.Host == "" {
		httpHelpers.LogRequest(s.base.Name, req, "Missing host")
		httpHelpers.WriteError(w, 400, "Missing host")
		return
	}

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
	if err != nil && err != io.EOF {
		panic(err)
	}
	if s.ShouldCache(req, res) {
		s.CacheAsync(req, res, body, logError)
	}
}

func (s *Server) ServeAsAPI(w http.ResponseWriter, req *http.Request) {
	httpHelpers.LogRequest(s.base.Name, req, "API access")
	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil && err != io.EOF {
		panic(err)
	}
}

func (s *Server) Fetch(req *http.Request) (*http.Response, error) {
	if s.HasCache(req) {
		httpHelpers.LogRequest(s.base.Name, req, "Cache access")
		return s.FetchFromCache(req)
	}
	httpHelpers.LogRequest(s.base.Name, req, "Proxy access")
	return s.FetchFromServer(req)
}

func (s *Server) HasCache(req *http.Request) bool {
	key := GetKeyForRequest(req)
	if s.RedisAvailable() {
		val, err := s.redis.Exists(key).Result()
		if err != nil {
			panic(err)
		}
		return val == 1
	} else {
		_, found := s.cache.Get(key)
		return found
	}
}

func (s *Server) FetchFromCache(req *http.Request) (*http.Response, error) {
	key := GetKeyForRequest(req)
	b, err := s.FetchRawFromCache(key)
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

func (s *Server) FetchRawFromCache(key string) ([]byte, error) {
	if s.RedisAvailable() {
		return s.redis.Get(key).Bytes()
	} else {
		b, found := s.cache.Get(key)
		if !found {
			return nil, fmt.Errorf("Cache with key '%s' not found", key)
		}
		return b.([]byte), nil
	}
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

func (s *Server) ShouldCache(req *http.Request, res *http.Response) bool {
	// TODO: Add logic on which request or response we should cache
	code := res.StatusCode
	return code == 200
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
	if s.RedisAvailable() {
		status := s.redis.Set(key, cache, DefaultExpirationTime)
		return status.Err()
	} else {
		s.cache.Set(key, cache, DefaultExpirationTime)
		return nil
	}
}

func (s *Server) RedisAvailable() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.redis != nil && s.redisAvailable
}

func (s *Server) serve(l net.Listener) {
	go s.startRedisHeartbeat()
	err := http.Serve(l, s)
	if err != nil {
		// do nothing
	}
}

func (s *Server) startRedisHeartbeat() {
	for s.Running() {
		redisAvailable := false
		if s.redis != nil {
			redisAvailable = s.checkRedisHeartbeat()
		}
		s.lock.Lock()
		s.redisAvailable = redisAvailable
		s.lock.Unlock()
		time.Sleep(DefaultHeartbeatTime)
	}
}

func (s *Server) checkRedisHeartbeat() bool {
	val, err := s.redis.Ping().Result()
	if err != nil {
		log.Printf("Redis Heartbeat: Got error '%s'", err.Error())
		return false
	}
	val = strings.TrimSpace(val)
	if val != "PONG" {
		log.Printf("Redis Heartbeat: Expected 'PONG' response, got '%s'", val)
		return false
	}
	log.Printf("Redis Heartbeat: %s", val)
	return true
}

func (c *CacheReader) Read(p []byte) (int, error) {
	return c.Reader.Read(p)
}

func (c *CacheReader) Close() error {
	return nil
}

func GetKeyForRequest(req *http.Request) string {
	return req.URL.Path
}

func logError(err error) {
	if err != nil {
		log.Println(err)
	}
}
