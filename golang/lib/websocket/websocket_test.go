package websocket

import (
	"io"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Frizz925/gbf-proxy/golang/lib/websocket/mocks"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type MockedConn struct {
	mocks.NetConn
}

type MockedWriter struct {
	io.Writer
	mock.Mock
}

type MockedReader struct {
	io.Reader
	mock.Mock
}

func (c *MockedConn) Write(p []byte) (int, error) {
	return len(p), nil
}

func TestWebsocket(t *testing.T) {
	conn := &MockedConn{}
	conn.On("Close").Return(nil)

	dialer := &websocket.Dialer{
		NetDial: func(string, string) (net.Conn, error) {
			return conn, nil
		},
	}

	c := NewController(&Config{
		Dialer: dialer,
		URL: &url.URL{
			Scheme: "ws",
			Host:   "localhost:8000",
			Path:   "/",
		},
		ErrorHandler: func(err error) {
			require.Nil(t, err)
		},
	})
	assert.Equal(t, NotConnectedError, c.Disconnect())
	assert.Equal(t, NotInitializedError, c.CheckLiveness())
	assert.Equal(t, NotInitializedError, c.Write(nil))
	_, err := c.Read()
	assert.Equal(t, NotInitializedError, err)

	assert.False(t, c.Connected())
	assert.False(t, c.Healthy())
	assert.Nil(t, c.Connect())
}
