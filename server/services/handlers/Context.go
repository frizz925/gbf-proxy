package handlers

import (
	"gbf-proxy/lib/logger"
	"net"
)

type Context struct {
	net.Conn
	logger.Logger
}
