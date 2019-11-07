package handlers

import (
	"net/http"
)

type ProxyHandler struct {
	*http.Client
}

var _ RequestHandler = (*ProxyHandler)(nil)

var DefaultHttpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func NewProxyHandler(clients ...*http.Client) *ProxyHandler {
	client := DefaultHttpClient
	if len(clients) > 0 {
		client = clients[0]
	}
	return &ProxyHandler{
		Client: client,
	}
}

func (h *ProxyHandler) HandleRequest(req *http.Request, ctx RequestContext) (*http.Response, error) {
	ctx.Logger.Info("Proxying request:", requestToString(req))
	res, err := h.Client.Do(outgoingRequest(req))
	if err != nil {
		return nil, err
	}
	return incomingResponse(res), nil
}

func outgoingRequest(req *http.Request) *http.Request {
	return &http.Request{
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Method:     req.Method,
		URL:        req.URL,
		Header:     req.Header,
		Host:       req.Host,
		Body:       req.Body,
	}
}

func incomingResponse(res *http.Response) *http.Response {
	return &http.Response{
		Proto:      res.Proto,
		ProtoMajor: res.ProtoMajor,
		ProtoMinor: res.ProtoMinor,
		Status:     res.Status,
		StatusCode: res.StatusCode,
		Header:     res.Header,
		Body:       res.Body,
	}
}
