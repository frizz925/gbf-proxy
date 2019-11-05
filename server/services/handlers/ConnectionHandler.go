package handlers

import (
	"gbf-proxy/lib/logger"
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
	connLogger := logger.NewConnectionLogger(conn, h.Logger)
	defer connLogger.Info("Connection closed")
	connLogger.Info("Connection accepted")
	err := h.StreamForwarder.Forward(Context{
		Conn:   conn,
		Logger: connLogger,
	}, conn, conn)
	if err != nil {
		connLogger.Error(err)
	}
	return nil
}
