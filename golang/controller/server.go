package controller

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/Frizz925/gbf-proxy/golang/lib"
)

type ServerConfig struct {
	WebAddr string
	WebHost string
}

type Server struct {
	base   *lib.BaseServer
	config *ServerConfig
}

func New(config *ServerConfig) lib.Server {
	return &Server{
		base:   lib.NewBaseServer("Controller"),
		config: config,
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

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Println(err)
			writeError(w, 503, err.Error())
		}
		req.Body.Close()
	}()
	s.ServeHTTPUnsafe(w, req)
}

func (s *Server) ServeHTTPUnsafe(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s %s %s", req.RemoteAddr, req.Method, req.RequestURI)
	u, err := url.Parse(req.RequestURI)
	if err != nil {
		writeError(w, 400, "Bad request URI")
		return
	}

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
	} else {
		writeError(w, 403, "Host not allowed")
		return
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}
	c := http.Client{}
	res, err := c.Do(&http.Request{
		Method: req.Method,
		URL:    u,
		Body:   req.Body,
		Header: req.Header,
	})
	if err != nil {
		writeServerError(w, 502, "Bad gateway", err)
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
	err := http.Serve(l, s)
	if err != nil {
		// do nothing
	}
}

func writeServerError(w http.ResponseWriter, code int, message string, err error) {
	log.Println(err)
	writeError(w, code, message)
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, err := w.Write([]byte(message + "\r\n"))
	if err != nil {
		panic(err)
	}
}
