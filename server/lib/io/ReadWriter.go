package io

import "io"

type ReadWriter struct {
	io.Reader
	io.Writer
}

func NewReadWriter(r io.Reader, w io.Writer) io.ReadWriter {
	return &ReadWriter{r, w}
}
