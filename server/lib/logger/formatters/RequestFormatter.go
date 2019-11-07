package formatters

import (
	"fmt"
	"net/http"
)

type RequestFormatter struct {
	Request *http.Request
}

var _ LogFormatter = (*RequestFormatter)(nil)

func NewRequestFormatter(req *http.Request) *RequestFormatter {
	return &RequestFormatter{
		Request: req,
	}
}

func (f *RequestFormatter) Format(message string) string {
	forwardedFor := f.Request.Header.Get("X-Forwarded-For")
	return fmt.Sprintf("[%s] %s", forwardedFor, message)
}
