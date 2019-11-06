package handlers

import (
	"bufio"
	"fmt"
	httplib "gbf-proxy/lib/http"
	"gbf-proxy/lib/logger"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const BUFFER_SIZE = 4096

type GatewayHandler struct {
	proxyHandler RequestHandler
	webHandler   RequestHandler
	pool         *sync.Pool
	hostCache    map[string]bool
	assetCache   map[string]bool
	log          logger.Logger
}

var _ StreamForwarder = (*GatewayHandler)(nil)
var _ RequestHandler = (*GatewayHandler)(nil)

func NewGatewayHandler(proxyHandler RequestHandler, webHandler RequestHandler) *GatewayHandler {
	return &GatewayHandler{
		proxyHandler: proxyHandler,
		webHandler:   webHandler,
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, BUFFER_SIZE)
			},
		},
		hostCache:  make(map[string]bool),
		assetCache: make(map[string]bool),
		log:        logger.DefaultLogger,
	}
}

func (h *GatewayHandler) Forward(r io.Reader, w io.Writer) error {
	reader := bufio.NewReader(r)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}
	return h.ForwardRequest(sanitizeRequest(req), reader, w)
}

func (h *GatewayHandler) ForwardRequest(req *http.Request, r *bufio.Reader, w io.Writer) error {
	reqStr := requestToString(req)
	if req.Method == "CONNECT" {
		h.log.Info("Responding to CONNECT request:", reqStr)
		err := h.respondConnect(req, w)
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
	if h.RequestAllowed(req) {
		if req.URL.Scheme != "http" || !h.AssetRequest(req) {
			h.log.Info("Tunneling request:", reqStr)
			return h.ForwardTunnel(req, req.URL, r, w)
		}
	}
	h.log.Info("Intercepting request:", reqStr)
	return h.ForwardIntercept(req, w)
}

func (h *GatewayHandler) ForwardIntercept(req *http.Request, w io.Writer) error {
	res, err := h.HandleRequest(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return res.Write(w)
}

func (h *GatewayHandler) ForwardTunnel(req *http.Request, u *url.URL, r io.Reader, w io.Writer) error {
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
		cerr <- copyStream(h.pool, conn, w)
	}()
	go func() {
		cerr <- copyStream(h.pool, r, conn)
	}()
	return <-cerr
}

func (h *GatewayHandler) HandleRequest(req *http.Request) (*http.Response, error) {
	reqStr := requestToString(req)
	if h.RequestAllowed(req) {
		h.log.Info("Directing request to proxy handler:", reqStr)
		return h.proxyHandler.HandleRequest(req)
	} else {
		h.log.Info("Directing request to web handler:", reqStr)
		return h.webHandler.HandleRequest(req)
	}
}

func (h *GatewayHandler) RequestAllowed(req *http.Request) bool {
	host := req.URL.Hostname()
	if v, ok := h.hostCache[host]; ok {
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
	h.hostCache[host] = true
	return true
}

func (h *GatewayHandler) AssetRequest(req *http.Request) bool {
	host := req.URL.Hostname()
	if v, ok := h.assetCache[host]; ok {
		return v
	} else if !strings.HasPrefix(host, "game-a") && !strings.HasPrefix(host, "gbf.game-a") {
		return false
	}
	h.assetCache[host] = true
	return true
}

func (h *GatewayHandler) respondConnect(req *http.Request, w io.Writer) error {
	return httplib.NewResponseBuilder(req).
		StatusCode(200).
		Status("200 Connection Established").
		Build().
		Write(w)
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
