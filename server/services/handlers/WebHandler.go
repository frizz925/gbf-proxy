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
	u := req.URL
	reqStr := requestToString(req)
	if u.Hostname() != h.hostname {
		ctx.Logger.Info("Denying access:", reqStr)
		return ForbiddenHostResponse(req), nil
	}
	forwardedScheme := req.Header.Get("X-Forwarded-Scheme")
	if forwardedScheme == "http" {
		u.Scheme = "https"
		u.Host = u.Hostname()
		ctx.Logger.Info("Redirecting to HTTPS site:", reqStr)
		return RedirectResponse(req, req.URL.String()), nil
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

func RedirectResponse(req *http.Request, location string) *http.Response {
	return httplib.NewResponseBuilder(req).
		StatusCode(301).
		Status("301 Moved Permanently").
		AddHeader("Location", location).
		Build()
}
