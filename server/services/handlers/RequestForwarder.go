package handlers

import (
	"net/http"
)

type RequestForwarder interface {
	ForwardRequest(Context, *http.Request) (*http.Response, error)
}
