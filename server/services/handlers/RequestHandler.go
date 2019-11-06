package handlers

import (
	"net/http"
)

type RequestHandler interface {
	HandleRequest(*http.Request) (*http.Response, error)
}
