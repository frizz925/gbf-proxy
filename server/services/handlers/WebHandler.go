package handlers

import (
	"fmt"
	httplib "gbf-proxy/lib/http"
	"gbf-proxy/lib/logger"
	"net/http"
)

type WebHandler struct {
	hostname string
	log      logger.Logger
}

var _ RequestHandler = (*WebHandler)(nil)

func NewWebHandler(hostname string) *WebHandler {
	return &WebHandler{
		hostname: hostname,
		log:      logger.DefaultLogger,
	}
}

func (h *WebHandler) HandleRequest(req *http.Request) (*http.Response, error) {
	reqStr := requestToString(req)
	if req.URL.Hostname() != h.hostname {
		h.log.Info("Denying tunneling attempt:", reqStr)
		return ForbiddenHostResponse(req), nil
	}
	if req.URL.Path == "/" {
		return RedirectResponse(req, "https://github.com/Frizz925/gbf-proxy"), nil
	}
	return StatusResponse(req, 404, "404 Not Found"), nil
}

func ForbiddenHostResponse(req *http.Request) *http.Response {
	host := req.URL.Hostname()
	message := fmt.Sprintf("Target host %s is not allowed to be accessed via this proxy", host)
	return httplib.NewResponseBuilder(req).
		StatusCode(403).
		Status("403 Forbidden").
		BodyString(message).
		Build()
}

func RedirectResponse(req *http.Request, location string) *http.Response {
	return httplib.NewResponseBuilder(req).
		StatusCode(302).
		Status("302 Found").
		AddHeader("Location", location).
		Build()
}

func StatusResponse(req *http.Request, statusCode int, status string) *http.Response {
	return httplib.NewResponseBuilder(req).
		StatusCode(statusCode).
		Status(status).
		Build()
}
