package handlers

import "io"

type StreamForwarder interface {
	Forward(io.Reader, io.Writer) error
}
