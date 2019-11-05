package services

import "net"

type ListeningService interface {
	Listen(net.Listener) error
}
