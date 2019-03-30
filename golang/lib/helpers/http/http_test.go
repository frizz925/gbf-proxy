package http

import (
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/Frizz925/gbf-proxy/golang/lib/helpers/http/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSerializeRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	body := mocks.NewMockReadCloser(ctrl)
	expectedError := errors.New("Read error")
	body.
		EXPECT().
		Read(gomock.Any()).
		Return(0, expectedError)

	req, err := SerializeRequest(&http.Request{
		Body: body,
	})
	assert.Nil(t, req)
	assert.Equal(t, expectedError, err)
}

func TestSerializeResponse(t *testing.T) {
	ctrl := gomock.NewController(t)
	body := mocks.NewMockReadCloser(ctrl)
	expectedError := errors.New("Read error")
	body.
		EXPECT().
		Read(gomock.Any()).
		Return(0, expectedError)

	res, err := SerializeResponse(&http.Response{
		Body: body,
	})
	assert.Nil(t, res)
	assert.Equal(t, expectedError, err)
}

func TestWriteError(t *testing.T) {
	ctrl := gomock.NewController(t)

	logger := mocks.NewMockLogger(ctrl)
	logger.
		EXPECT().
		Error(gomock.Any()).
		Return().
		Times(1)

	responseWriter := mocks.NewMockResponseWriter(ctrl)
	expectedError := errors.New("Write error")
	responseWriter.
		EXPECT().
		WriteHeader(gomock.Any()).
		Return()
	responseWriter.
		EXPECT().
		Write(gomock.Any()).
		Return(0, expectedError)

	WriteError(logger, responseWriter, 500, "Generic error")
}

func TestAddrToHost(t *testing.T) {
	assert.Equal(t, "localhost", AddrToHost("localhost"))
	assert.Equal(t, "127.0.0.1", AddrToHost("127.0.0.1:8000"))
}
