package io

import (
	"io"
	"sync"
)

const BUFFER_SIZE = 4096

var bufferPool *sync.Pool

func init() {
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, BUFFER_SIZE)
		},
	}
}

func DuplexStream(dst io.ReadWriter, src io.ReadWriter) error {
	c := make(chan error, 1)
	go func(cw chan<- error) {
		cw <- Stream(dst, src)
	}(c)
	go func(cw chan<- error) {
		cw <- Stream(src, dst)
	}(c)
	return <-c
}

func Stream(r io.Reader, w io.Writer) error {
	b := GetBuffer()
	defer PutBuffer(b)
	_, err := io.CopyBuffer(w, r, b)
	return err
}

func GetBuffer() []byte {
	return bufferPool.Get().([]byte)
}

func PutBuffer(b []byte) {
	bufferPool.Put(b)
}
