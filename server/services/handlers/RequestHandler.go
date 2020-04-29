package handlers

import (
	"net/http"
)

type RequestHandler interface {
	HandleRequest(*http.Request, *RequestContext) (*http.Response, error)
}
