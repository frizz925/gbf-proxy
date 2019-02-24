package http

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Frizz925/gbf-proxy/golang/lib/logging"
)

func WriteServerError(w http.ResponseWriter, code int, message string, err error) {
	log.Println(err)
	WriteError(w, code, message)
}

func WriteError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	_, err := w.Write([]byte(message + "\r\n"))
	if err != nil {
		log.Println(err)
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

func LogRequest(logger *logging.Logger, req *http.Request, message string) {
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
