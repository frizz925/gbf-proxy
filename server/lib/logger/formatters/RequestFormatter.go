package formatters

import (
	"fmt"
	"net"
	"net/http"
	"strings"
)

type RequestFormatter struct {
	net.Conn
	*http.Request
}

var _ LogFormatter = (*RequestFormatter)(nil)

func NewRequestFormatter(conn net.Conn, req *http.Request) *RequestFormatter {
	return &RequestFormatter{
		Conn:    conn,
		Request: req,
	}
}

func (f *RequestFormatter) Format(prefix string, message string) (string, string) {
	sourceAddr := f.Request.Header.Get("X-Forwarded-For")
	if sourceAddr == "" {
		remoteAddr := f.Conn.RemoteAddr().String()
		idx := strings.Index(remoteAddr, ":")
		if idx > 0 {
			sourceAddr = remoteAddr[:idx]
		} else {
			sourceAddr = remoteAddr
		}
	}
	return fmt.Sprintf("[%-15s] %s", sourceAddr, prefix), message
}
