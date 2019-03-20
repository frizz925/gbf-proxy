package http

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/Frizz925/gbf-proxy/golang/lib/logging"
)

type Base struct {
	Header        http.Header
	ContentLength int64
	Body          []byte
}

type Request struct {
	Base   Base
	Method string
	URL    url.URL
	Host   string
}

type Response struct {
	Base       Base
	Status     string
	StatusCode int
}

type StubCloserReader struct {
	Reader io.Reader
}

func NewBodyReader(body []byte) io.ReadCloser {
	return &StubCloserReader{
		Reader: bytes.NewReader(body),
	}
}

func (r *StubCloserReader) Read(p []byte) (int, error) {
	return r.Reader.Read(p)
}

func (r *StubCloserReader) Close() error {
	return nil
}

func SerializeRequest(req *http.Request) (*Request, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return &Request{
		Base: Base{
			Header:        req.Header,
			ContentLength: req.ContentLength,
			Body:          body,
		},
		Method: req.Method,
		URL:    *req.URL,
		Host:   req.Host,
	}, nil
}

func SerializeResponse(res *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return &Response{
		Base: Base{
			Header:        res.Header,
			ContentLength: res.ContentLength,
			Body:          body,
		},
		Status:     res.Status,
		StatusCode: res.StatusCode,
	}, nil
}

func UnserializeRequest(req *Request) (*http.Request, error) {
	return &http.Request{
		Method:        req.Method,
		URL:           &req.URL,
		Header:        req.Base.Header,
		ContentLength: req.Base.ContentLength,
		Body:          NewBodyReader(req.Base.Body),
	}, nil
}

func UnserializeResponse(res *Response) (*http.Response, error) {
	return &http.Response{
		Status:        res.Status,
		StatusCode:    res.StatusCode,
		Header:        res.Base.Header,
		ContentLength: res.Base.ContentLength,
		Body:          NewBodyReader(res.Base.Body),
	}, nil
}

func WriteServerError(logger logging.Logger, w http.ResponseWriter, code int, message string, err error) {
	logger.Error(err)
	WriteError(logger, w, code, message)
}

func WriteError(logger logging.Logger, w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, err := w.Write([]byte(message + "\r\n"))
	if err != nil {
		logger.Error(err)
	}
}

func ParseURL(req *http.Request) *url.URL {
	u := req.URL
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		u.Host = req.Host
	}
	return u
}

func LogRequest(logger logging.Logger, req *http.Request, message string) {
	u := ParseURL(req)
	logger.Infof("%s %s %s - %s", req.RemoteAddr, req.Method, u.String(), message)
}

func AddrToHost(addr string) string {
	tokens := strings.SplitN(addr, ":", 2)
	if len(tokens) >= 2 {
		return tokens[0]
	}
	return addr
}
