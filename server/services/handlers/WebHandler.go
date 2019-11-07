package handlers

import (
	"fmt"
	httplib "gbf-proxy/lib/http"
	"net/http"
)

type WebHandler struct {
	hostname string
	remote   *RemoteHandler
}

var _ RequestHandler = (*WebHandler)(nil)

func NewWebHandler(hostname string, addr string) *WebHandler {
	return &WebHandler{
		hostname: hostname,
		remote:   NewRemoteHandler(addr),
	}
}

func (h *WebHandler) HandleRequest(req *http.Request, ctx RequestContext) (*http.Response, error) {
	reqStr := requestToString(req)
	if req.URL.Hostname() != h.hostname {
		ctx.Logger.Info("Denying access:", reqStr)
		return ForbiddenHostResponse(req), nil
	}
	return h.remote.HandleRequest(req, ctx)
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
