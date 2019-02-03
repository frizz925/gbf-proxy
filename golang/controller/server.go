package controller

import (
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
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

func (s *Server) Open(addr string) (net.Listener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s.Listener = l
	s.WaitGroup.Add(1)
	go s.serve()
	return l, nil
}

func (s *Server) Close() error {
	if s.Listener == nil {
		return errors.New("Server isn't running")
	}
	s.Listener.Close()
	s.WaitGroup.Wait()
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	host := req.Host
	hostname := host
	tokens := strings.SplitN(host, ":", 2)
	if len(tokens) >= 2 {
		hostname = tokens[0]
	}
	if !strings.HasSuffix(hostname, ".granbluefantasy.jp") {
		writeError(w, 403, "Host not allowed")
		return
	}

	url, err := url.Parse(req.RequestURI)
	if err != nil {
		writeError(w, 400, "Bad request URI")
		return
	}
	url.Host = host

	c := http.Client{}
	res, err := c.Do(&http.Request{
		Method: req.Method,
		URL:    req.URL,
		Body:   req.Body,
		Header: req.Header,
	})
	if err != nil {
		panic(err)
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
	for written := 0; written < len(body); {
		write, err := w.Write(body)
		if err != nil {
			panic(err)
		}
		written += write
	}
}

func (s *Server) serve() {
	defer s.close()
	err := http.Serve(s.Listener, s)
	if err != nil {
		// do nothing
	}
}

func (s *Server) close() {
	s.Listener.Close()
	s.WaitGroup.Done()
}

func writeError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, err := w.Write([]byte(message))
	if err != nil {
		panic(err)
	}
}
