package http

import (
	"log"
	"net/http"
	"net/url"
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

func LogRequest(name string, req *http.Request, message string) {
	u := ParseURL(req)
	log.Printf("[%s] %s %s %s - %s", name, req.RemoteAddr, req.Method, u.String(), message)
}
