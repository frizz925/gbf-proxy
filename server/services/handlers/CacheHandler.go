package handlers

import (
	"bytes"
	"fmt"
	"gbf-proxy/lib/cache"
	"gbf-proxy/lib/logger"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type CacheHandler struct {
	handler   RequestHandler
	cache     cache.Client
	hostCache map[string]bool
}

type CacheContext struct {
	handler   RequestHandler
	cache     cache.Client
	hostCache map[string]bool
	log       *logger.Logger
}

type cachedResponse struct {
	Proto            string
	ProtoMajor       int
	ProtoMinor       int
	Status           string
	StatusCode       int
	Header           http.Header
	Body             []byte
	ContentLength    int64
	TransferEncoding []string
	Uncompressed     bool
	Trailer          http.Header
}

var _ RequestHandler = (*CacheHandler)(nil)

func NewCacheHandler(rh RequestHandler, c cache.Client) *CacheHandler {
	return &CacheHandler{
		handler:   rh,
		cache:     c,
		hostCache: make(map[string]bool),
	}
}

func (h *CacheHandler) HandleRequest(req *http.Request, ctx RequestContext) (*http.Response, error) {
	return (CacheContext{
		handler:   h.handler,
		cache:     h.cache,
		hostCache: h.hostCache,
		log:       ctx.Logger,
	}).HandleRequest(req, ctx)
}

func (c CacheContext) HandleRequest(req *http.Request, ctx RequestContext) (*http.Response, error) {
	if !c.shouldCacheRequest(req) {
		return c.handler.HandleRequest(req, ctx)
	}

	key := c.getCacheKey(req.URL)
	exists, err := c.cache.Has(key)
	if err != nil {
		c.log.Error("Cache ERROR:", err)
	} else if !exists {
		c.log.Info("Cache MISS:", key)
	} else {
		c.log.Info("Cache HIT:", key)
		return c.getCache(key, req)
	}

	res, err := c.handler.HandleRequest(req, ctx)
	if err != nil {
		return nil, err
	}
	if !c.shouldCacheResponse(res) {
		return res, nil
	}
	return c.putCacheAsync(key, req, res)
}

func (c CacheContext) shouldCacheRequest(req *http.Request) bool {
	if req.Method != "GET" {
		return false
	}

	host := req.URL.Hostname()
	if v, ok := c.hostCache[host]; ok {
		return v
	} else if strings.HasPrefix(host, "game-a") && strings.HasSuffix(host, ".granbluefantasy.jp") {
		// do nothing
	} else if strings.HasPrefix(host, "gbf.game-a") && strings.HasSuffix(host, ".mbga.jp") {
		// do nothing
	} else {
		return false
	}
	c.hostCache[host] = true
	return true
}

func (c CacheContext) shouldCacheResponse(res *http.Response) bool {
	return res.StatusCode >= 200 && res.StatusCode < 300
}

func (c CacheContext) getCache(key string, req *http.Request) (*http.Response, error) {
	cr := &cachedResponse{}
	err := c.cache.Get(key, cr)
	if err != nil {
		return nil, err
	}
	return cr.unmarshal(req), nil
}

func (c CacheContext) putCacheAsync(key string, req *http.Request, res *http.Response) (*http.Response, error) {
	cr, err := marshalResponse(res)
	if err != nil {
		return nil, err
	}
	go func() {
		err := c.putCache(key, cr)
		if err != nil {
			c.log.Error(err)
		}
	}()
	return cr.unmarshal(req), nil
}

func (c CacheContext) putCache(key string, cr *cachedResponse) error {
	err := c.cache.Set(key, cr)
	if err != nil {
		return err
	}
	c.log.Infof("Cache PUT: %s", key)
	return nil
}

func (c CacheContext) getCacheKey(u *url.URL) string {
	query := u.RawQuery
	if query == "" {
		return u.Path
	}
	return fmt.Sprintf("%s.%s", u.Path, query)
}

func marshalResponse(res *http.Response) (*cachedResponse, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return &cachedResponse{
		Proto:            res.Proto,
		ProtoMajor:       res.ProtoMajor,
		ProtoMinor:       res.ProtoMinor,
		Status:           res.Status,
		StatusCode:       res.StatusCode,
		Header:           res.Header,
		Body:             body,
		ContentLength:    res.ContentLength,
		TransferEncoding: res.TransferEncoding,
		Uncompressed:     res.Uncompressed,
		Trailer:          res.Trailer,
	}, nil
}

func (c *cachedResponse) unmarshal(req *http.Request) *http.Response {
	header := c.Header
	aclOrigin := header.Get("Access-Control-Allow-Origin")
	if aclOrigin != "" {
		reqOrigin := req.Header.Get("Origin")
		if reqOrigin == "" {
			reqOrigin = "*"
		}
		header.Set("Access-Control-Allow-Origin", reqOrigin)
	}
	return &http.Response{
		Proto:            c.Proto,
		ProtoMajor:       c.ProtoMajor,
		ProtoMinor:       c.ProtoMinor,
		Status:           c.Status,
		StatusCode:       c.StatusCode,
		Header:           header,
		Body:             c.newReader(),
		ContentLength:    c.ContentLength,
		TransferEncoding: c.TransferEncoding,
		Uncompressed:     c.Uncompressed,
		Trailer:          c.Trailer,
		Request:          req,
	}
}

func (c *cachedResponse) newReader() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(c.Body))
}
