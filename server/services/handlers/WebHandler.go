package handlers

import (
	"fmt"
	httplib "gbf-proxy/lib/http"
	"gbf-proxy/lib/logger"
	"net/http"
)

type WebHandler struct {
	hostname string
	remote   *RemoteHandler
	log      logger.Logger
}

var _ RequestHandler = (*WebHandler)(nil)

func NewWebHandler(hostname string, addr string) *WebHandler {
	return &WebHandler{
		hostname: hostname,
		remote:   NewRemoteHandler(addr),
		log:      logger.DefaultLogger,
	}
}

func (h *WebHandler) HandleRequest(req *http.Request) (*http.Response, error) {
	reqStr := requestToString(req)
	if req.URL.Hostname() != h.hostname {
		h.log.Info("Denying access:", reqStr)
		return ForbiddenHostResponse(req), nil
	}
	return h.remote.HandleRequest(req)
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
