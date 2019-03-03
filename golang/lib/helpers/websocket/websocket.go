package websocket

import (
	httpHelpers "github.com/Frizz925/gbf-proxy/golang/lib/helpers/http"
)

type Request struct {
	ID      string
	Payload httpHelpers.Request
}

type Response struct {
	ID      string
	Payload httpHelpers.Response
}
