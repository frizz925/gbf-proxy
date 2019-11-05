package handlers

import "net"

type ConnectionForwarder interface {
	ForwardConnection(net.Conn) error
}
