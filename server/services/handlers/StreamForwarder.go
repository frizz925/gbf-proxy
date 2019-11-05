package handlers

import "io"

type StreamForwarder interface {
	Forward(Context, io.Reader, io.Writer) error
}
