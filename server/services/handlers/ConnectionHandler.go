package handlers

import (
	"gbf-proxy/lib/logger"
	"io"
	"net"
)

type ConnectionHandler struct {
	StreamForwarder
	logger.Logger
}

var _ ConnectionForwarder = (*ConnectionHandler)(nil)

func NewConnectionHandler(sf StreamForwarder) *ConnectionHandler {
	return &ConnectionHandler{
		StreamForwarder: sf,
		Logger:          logger.Factory.New(1),
	}
}

func (h *ConnectionHandler) ForwardConnection(conn net.Conn) error {
	err := h.StreamForwarder.Forward(Context{
		Conn:   conn,
		Logger: h.Logger,
	}, conn, conn)
	if err != nil && err != io.EOF {
		h.Logger.Error(err)
	}
	return nil
}
