package handlers

import (
	"bufio"
	"fmt"
	connlib "gbf-proxy/lib/conn"
	httplib "gbf-proxy/lib/http"
	iolib "gbf-proxy/lib/io"
	"gbf-proxy/lib/logger"
	"gbf-proxy/lib/logger/formatters"
	"io"
	"net/http"
	"strings"
	"sync"
)

type GatewayHandler struct {
	version      string
	proxyHandler RequestHandler
	webHandler   RequestHandler
	pool         *sync.Pool
	hostCache    map[string]bool
	assetCache   map[string]bool
}

var _ StreamForwarder = (*GatewayHandler)(nil)
var _ RequestHandler = (*GatewayHandler)(nil)

func NewGatewayHandler(version string, proxyHandler RequestHandler, webHandler RequestHandler) *GatewayHandler {
	return &GatewayHandler{
		version:      version,
		proxyHandler: proxyHandler,
		webHandler:   webHandler,
		hostCache:    make(map[string]bool),
		assetCache:   make(map[string]bool),
	}
}

func (h *GatewayHandler) Forward(r io.Reader, w io.Writer) error {
	reader := bufio.NewReader(r)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}
	req = sanitizeRequest(req)
	defer req.Body.Close()
	return h.ForwardRequest(req, RequestContext{
		Logger: h.CreateRequestLogger(req),
	}, reader, w)
}

func (h *GatewayHandler) ForwardRequest(req *http.Request, ctx RequestContext, r *bufio.Reader, w io.Writer) error {
	reqStr := requestToString(req)
	if req.Method == "CONNECT" {
		ctx.Logger.Info("Responding to CONNECT request:", reqStr)
		if !h.RequestAllowed(req) {
			ctx.Logger.Info("Denying CONNECT request:", reqStr)
			return h.respondForbidden(req, w)
		}
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
			defer req.Body.Close()
		}
	}
	if h.RequestAllowed(req) {
		if req.URL.Scheme != "http" || !h.AssetRequest(req) {
			ctx.Logger.Info("Tunnelling request:", reqStr)
			return h.ForwardTunnel(req, r, w)
		}
	}
	ctx.Logger.Info("Intercepting request:", reqStr)
	return h.ForwardIntercept(req, ctx, w)
}

func (h *GatewayHandler) ForwardIntercept(req *http.Request, ctx RequestContext, w io.Writer) error {
	res, err := h.HandleRequest(req, ctx)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return res.Write(w)
}

func (h *GatewayHandler) ForwardTunnel(req *http.Request, r io.Reader, w io.Writer) error {
	u := req.URL
	conn, err := connlib.CreateURLConnection(u)
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
	return iolib.DuplexStream(conn, iolib.NewReadWriter(r, w))
}

func (h *GatewayHandler) HandleRequest(req *http.Request, ctx RequestContext) (*http.Response, error) {
	reqStr := requestToString(req)
	if h.RequestAllowed(req) {
		ctx.Logger.Info("Directing request to proxy handler:", reqStr)
		return h.proxyHandler.HandleRequest(req, ctx)
	} else {
		ctx.Logger.Info("Directing request to web handler:", reqStr)
		return h.webHandler.HandleRequest(req, ctx)
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

func (h *GatewayHandler) CreateRequestLogger(req *http.Request) *logger.Logger {
	return &logger.Logger{
		Printers: logger.DefaultPrinters,
		Formatters: []formatters.LogFormatter{
			formatters.NewCallerFormatter(),
			formatters.NewRequestFormatter(req),
		},
	}
}

func (h *GatewayHandler) respondForbidden(req *http.Request, w io.Writer) error {
	host := req.URL.Hostname()
	message := fmt.Sprintf("Connection tunelling to host %s is not allowed", host)
	return httplib.NewResponseBuilder(req, h.version).
		StatusCode(403).
		Status("403 Forbidden").
		BodyString(message).
		Build().
		Write(w)
}

func (h *GatewayHandler) respondConnect(req *http.Request, w io.Writer) error {
	return httplib.NewResponseBuilder(req, h.version).
		StatusCode(200).
		Status("200 Connection Established").
		Build().
		Write(w)
}
