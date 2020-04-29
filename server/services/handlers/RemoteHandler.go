package handlers

import (
	"bufio"
	connlib "gbf-proxy/lib/conn"
	iolib "gbf-proxy/lib/io"
	"io"
	"net"
	"net/http"
)

type RemoteHandler struct {
	addr string
}

var _ RequestHandler = (*RemoteHandler)(nil)
var _ StreamForwarder = (*RemoteHandler)(nil)

func NewRemoteHandler(addr string) *RemoteHandler {
	return &RemoteHandler{
		addr,
	}
}

func (h *RemoteHandler) HandleRequest(req *http.Request, ctx *RequestContext) (*http.Response, error) {
	conn, err := h.CreateConnection()
	if err != nil {
		return nil, err
	}
	err = req.Write(conn)
	if err != nil {
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(conn), req)
}

func (h *RemoteHandler) Forward(r io.Reader, w io.Writer) error {
	conn, err := h.CreateConnection()
	if err != nil {
		return err
	}
	return iolib.DuplexStream(conn, iolib.NewReadWriter(r, w))
}

func (h *RemoteHandler) CreateConnection() (net.Conn, error) {
	return connlib.CreateConnection(h.addr)
}
