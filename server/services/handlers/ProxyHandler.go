package handlers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"gbf-proxy/lib/logger"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
)

const BUFFER_SIZE = 4096

type ProxyHandler struct {
	RequestForwarder
	*sync.Pool
	hostCache  map[string]bool
	assetCache map[string]bool
}

type ProxyContext struct {
	Context   Context
	Handler   *ProxyHandler
	Forwarder RequestForwarder
	Reader    *bufio.Reader
	Writer    io.Writer
	log       logger.Logger
}

var _ StreamForwarder = (*ProxyHandler)(nil)

func NewProxyHandler(rf RequestForwarder) *ProxyHandler {
	return &ProxyHandler{
		RequestForwarder: rf,
		Pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, BUFFER_SIZE)
			},
		},
		hostCache:  make(map[string]bool),
		assetCache: make(map[string]bool),
	}
}

func (h *ProxyHandler) Forward(ctx Context, r io.Reader, w io.Writer) error {
	reader := bufio.NewReader(r)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}
	return (&ProxyContext{
		Context:   ctx,
		Handler:   h,
		Forwarder: h.RequestForwarder,
		Reader:    reader,
		Writer:    w,
		log:       ctx.Logger,
	}).Forward(sanitizeRequest(req))
}

func (c *ProxyContext) Forward(req *http.Request) error {
	var (
		r = c.Reader
		w = c.Writer
	)
	reqStr := fmt.Sprintf("%s %s", req.Method, req.URL.String())
	if !c.requestAllowed(req) {
		c.log.Info("Denying request:", reqStr)
		return c.respondForbidden(req, w)
	}
	if req.Method == "CONNECT" {
		err := c.respondConnect(req, w)
		if err != nil {
			return err
		}
		if req.URL.Scheme == "http" {
			nextReq, err := http.ReadRequest(r)
			if err != nil {
				return err
			}
			req = sanitizeRequest(mergeRequests(req, nextReq))
		}
	}
	if req.URL.Scheme != "http" || !c.isAssetRequest(req) {
		c.log.Info("Tunneling request:", reqStr)
		return c.ForwardTunnel(req, req.URL)
	}
	c.log.Info("Intercepting request:", reqStr)
	return c.ForwardIntercept(req, r)
}

func (c *ProxyContext) ForwardIntercept(req *http.Request, reader *bufio.Reader) error {
	keepAlive := strings.ToLower(req.Header.Get("Connection")) == "keep-alive"
	nextReq := req
	for {
		res, err := c.Forwarder.ForwardRequest(c.Context, nextReq)
		if err != nil {
			return err
		}
		err = res.Write(c.Writer)
		res.Body.Close()
		if err != nil {
			return err
		}
		if !keepAlive {
			break
		}
		nextReq, err = http.ReadRequest(reader)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *ProxyContext) ForwardTunnel(req *http.Request, u *url.URL) error {
	conn, err := createConnection(req.URL, u.Scheme)
	if err != nil {
		return err
	}
	defer conn.Close()
	if u.Scheme == "http" {
		err = req.Write(conn)
		if err != nil {
			return err
		}
	}
	cerr := make(chan error, 1)
	go func() {
		cerr <- copyStream(c.Handler.Pool, conn, c.Writer)
	}()
	go func() {
		cerr <- copyStream(c.Handler.Pool, c.Reader, conn)
	}()
	return <-cerr
}

func (c *ProxyContext) requestAllowed(req *http.Request) bool {
	host := req.URL.Hostname()
	if v, ok := c.Handler.hostCache[host]; ok {
		return v
	} else if strings.HasPrefix(host, "game") && strings.HasSuffix(host, ".granbluefantasy.jp") {
		// do nothing
	} else if strings.HasPrefix(host, "gbf.game") && strings.HasSuffix(host, ".mbga.jp") {
		// do nothing
	} else if strings.HasSuffix(host, ".mobage.jp") {
		// do nothing
	} else {
		return false
	}
	c.Handler.hostCache[host] = true
	return true
}

func (c *ProxyContext) isAssetRequest(req *http.Request) bool {
	host := req.URL.Hostname()
	if v, ok := c.Handler.assetCache[host]; ok {
		return v
	} else if strings.HasPrefix(host, "game-a") || strings.HasPrefix(host, "gbf.game-a") {
		// do nothing
	} else {
		return false
	}
	c.Handler.assetCache[host] = true
	return true
}

func (c *ProxyContext) respondConnect(req *http.Request, w io.Writer) error {
	return c.respondRequest(req, w, "200 Connection Established", 200)
}

func (c *ProxyContext) respondForbidden(req *http.Request, w io.Writer) error {
	host := req.URL.Hostname()
	message := fmt.Sprintf("Host %s is not allowed to be accessed from this proxy\r\n", host)
	return c.respondRequest(req, w, "403 Forbidden", 200, message)
}

func (c *ProxyContext) respondRequest(req *http.Request, w io.Writer, status string, code int, content ...interface{}) error {
	buf := &bytes.Buffer{}
	for _, c := range content {
		if s, ok := c.(string); ok {
			buf.WriteString(s)
		} else if b, ok := c.([]byte); ok {
			buf.Write(b)
		} else {
			t := reflect.TypeOf(c)
			msg := fmt.Sprintf("Unrecognized body content of type %s", t.String())
			return errors.New(msg)
		}
	}
	return (&http.Response{
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Status:     status,
		StatusCode: code,
		Header:     c.createHeader(),
		Body:       ioutil.NopCloser(buf),
	}).Write(w)
}

func (c *ProxyContext) createHeader() http.Header {
	header := make(http.Header)
	header.Add("X-Proxy-Server", "Granblue Proxy")
	return header
}

func createConnection(u *url.URL, scheme string) (net.Conn, error) {
	addr := getAddress(u, scheme)
	return net.Dial("tcp4", addr)
}

func getAddress(u *url.URL, scheme string) string {
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	return fmt.Sprintf("%s:%s", host, port)
}

func copyStream(pool *sync.Pool, r io.Reader, w io.Writer) error {
	b := pool.Get().([]byte)
	defer pool.Put(b)
	_, err := io.CopyBuffer(w, r, b)
	return err
}
