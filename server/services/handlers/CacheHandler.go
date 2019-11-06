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
	RequestHandler
	cache.Client
	hostCache map[string]bool
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

var _ RequestHandler = (*CacheHandler)(nil)

func NewCacheHandler(rh RequestHandler, c cache.Client) *CacheHandler {
	return &CacheHandler{
		RequestHandler: rh,
		Client:         c,
		hostCache:      make(map[string]bool),
		log:            logger.DefaultLogger,
	}
}

func (h *CacheHandler) HandleRequest(req *http.Request) (*http.Response, error) {
	if !h.shouldCacheRequest(req) {
		return h.RequestHandler.HandleRequest(req)
	}

	key := h.getCacheKey(req.URL)
	exists, err := h.Client.Has(key)
	if err != nil {
		return nil, err
	}
	if exists {
		h.log.Infof("Cache HIT: %s", key)
		return h.getCache(key, req)
	}

	h.log.Infof("Cache MISS: %s", key)
	res, err := h.RequestHandler.HandleRequest(req)
	if err != nil {
		return nil, err
	}
	if !h.shouldCacheResponse(res) {
		return res, nil
	}
	return h.putCacheAsync(key, req, res)
}

func (h *CacheHandler) shouldCacheRequest(req *http.Request) bool {
	if req.Method != "GET" {
		return false
	}

	host := req.URL.Hostname()
	if v, ok := h.hostCache[host]; ok {
		return v
	} else if strings.HasPrefix(host, "game-a") && strings.HasSuffix(host, ".granbluefantasy.jp") {
		// do nothing
	} else if strings.HasPrefix(host, "gbf.game-a") && strings.HasSuffix(host, ".mbga.jp") {
		// do nothing
	} else {
		return false
	}
	h.hostCache[host] = true
	return true
}

func (h *CacheHandler) shouldCacheResponse(res *http.Response) bool {
	return res.StatusCode >= 200 && res.StatusCode < 300
}

func (h *CacheHandler) getCache(key string, req *http.Request) (*http.Response, error) {
	cr := &cachedResponse{}
	err := h.Client.Get(key, cr)
	if err != nil {
		return nil, err
	}
	return cr.unmarshal(req), nil
}

func (h *CacheHandler) putCacheAsync(key string, req *http.Request, res *http.Response) (*http.Response, error) {
	cr, err := marshalResponse(res)
	if err != nil {
		return nil, err
	}
	go func() {
		err := h.putCache(key, cr)
		if err != nil {
			h.log.Error(err)
		}
	}()
	return cr.unmarshal(req), nil
}

func (h *CacheHandler) putCache(key string, cr *cachedResponse) error {
	err := h.Client.Set(key, cr)
	if err != nil {
		return err
	}
	h.log.Infof("Cache PUT: %s", key)
	return nil
}

func (h *CacheHandler) getCacheKey(u *url.URL) string {
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
