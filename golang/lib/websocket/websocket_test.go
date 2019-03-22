package websocket

import (
	"bufio"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/Frizz925/gbf-proxy/golang/lib/websocket/mocks"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type MockedWriter struct {
	io.Writer
	mock.Mock
}

type MockedReader struct {
	io.Reader
	mock.Mock
}

func TestWebsocket(t *testing.T) {
	header := make(http.Header)
	header.Add("Connection", "upgrade")
	header.Add("Upgrade", "websocket")
	header.Add("Sec-WebSocket-Version", "13")
	header.Add("Sec-WebSocket-Key", "dGhlIHNhbXBsZSBub25jZQ==")

	request := &http.Request{
		Method: "GET",
		Header: header,
	}

	writer := &MockedWriter{}
	reader := &MockedReader{}
	readWriter := bufio.NewReadWriter(
		bufio.NewReader(reader),
		bufio.NewWriter(writer),
	)

	conn := &mocks.NetConn{}
	conn.On("SetDeadline", time.Time{}).Return(nil)
	conn.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	responseWriter := &mocks.ResponseWriter{}
	responseWriter.On("Header").Return(header)
	responseWriter.On("WriteHeader", mock.AnythingOfType("int")).Return()
	responseWriter.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)
	responseWriter.On("Hijack").Return(conn, readWriter, nil)

	upgrader := &websocket.Upgrader{}
	ws, err := upgrader.Upgrade(responseWriter, request, header)
	require.Nil(t, err)
	require.NotNil(t, ws)

	ctx := NewContext(ws)
	ctx.Init(func(err error) {
		require.NotNil(t, err)
	})
	ctx.Lock()
	ctx.Unlock()
}
