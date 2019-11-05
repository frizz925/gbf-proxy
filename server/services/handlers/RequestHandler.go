package handlers

import (
	"net/http"
)

type RequestHandler struct {
	*http.Client
}

var _ RequestForwarder = (*RequestHandler)(nil)

var DefaultHttpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func NewRequestHandler(clients ...*http.Client) *RequestHandler {
	client := DefaultHttpClient
	if len(clients) > 0 {
		client = clients[0]
	}
	return &RequestHandler{
		Client: client,
	}
}

func (h *RequestHandler) ForwardRequest(ctx Context, req *http.Request) (*http.Response, error) {
	ctx.Logger.Infof("Forwarding request: %s %s", req.Method, req.URL.String())
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
