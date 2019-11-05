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
	RequestForwarder
	cache.Client
	hostCache map[string]bool
}

type CacheContext struct {
	Context   Context
	Handler   *CacheHandler
	Forwarder RequestForwarder
	Cache     cache.Client
	log       logger.Logger
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

var _ RequestForwarder = (*CacheHandler)(nil)

func NewCacheHandler(rf RequestForwarder, c cache.Client) *CacheHandler {
	return &CacheHandler{
		RequestForwarder: rf,
		Client:           c,
		hostCache:        make(map[string]bool),
	}
}

func (h *CacheHandler) ForwardRequest(ctx Context, req *http.Request) (*http.Response, error) {
	return (&CacheContext{
		Context:   ctx,
		Handler:   h,
		Forwarder: h.RequestForwarder,
		Cache:     h.Client,
		log:       ctx.Logger,
	}).ForwardRequest(req)
}

func (c *CacheContext) ForwardRequest(req *http.Request) (*http.Response, error) {
	host := req.URL.Host
	if !c.shouldCache(host) {
		return c.Forwarder.ForwardRequest(c.Context, req)
	}
	key := c.getCacheKey(req.URL)
	exists, err := c.Cache.Has(key)
	if err != nil {
		return nil, err
	}
	if exists {
		c.log.Infof("Cache HIT: %s", key)
		return c.getCache(key, req)
	}
	c.log.Infof("Cache MISS: %s", key)
	res, err := c.Forwarder.ForwardRequest(c.Context, req)
	if err != nil {
		return nil, err
	}
	return c.putCacheAsync(key, req, res)
}

func (c *CacheContext) shouldCache(host string) bool {
	if v, ok := c.Handler.hostCache[host]; ok {
		return v
	} else if strings.HasPrefix(host, "game-a") && strings.HasSuffix(host, ".granbluefantasy.jp") {
		// do nothing
	} else if strings.HasPrefix(host, "gbf.game-a") && strings.HasSuffix(host, ".mbga.jp") {
		// do nothing
	} else {
		return false
	}
	c.Handler.hostCache[host] = true
	return true
}

func (c *CacheContext) getCache(key string, req *http.Request) (*http.Response, error) {
	cr := &cachedResponse{}
	err := c.Cache.Get(key, cr)
	if err != nil {
		return nil, err
	}
	return cr.unmarshal(req), nil
}

func (c *CacheContext) putCacheAsync(key string, req *http.Request, res *http.Response) (*http.Response, error) {
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

func (c *CacheContext) putCache(key string, cr *cachedResponse) error {
	err := c.Cache.Set(key, cr)
	if err != nil {
		return err
	}
	c.log.Infof("Cache PUT: %s", key)
	return nil
}

func (c *CacheContext) getCacheKey(u *url.URL) string {
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
	return &http.Response{
		Proto:            c.Proto,
		ProtoMajor:       c.ProtoMajor,
		ProtoMinor:       c.ProtoMinor,
		Status:           c.Status,
		StatusCode:       c.StatusCode,
		Header:           c.Header,
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
