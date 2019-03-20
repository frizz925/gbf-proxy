package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Frizz925/gbf-proxy/golang/lib/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockedReader struct {
	StubCloserReader
	mock.Mock
}

func (r *MockedReader) Read(p []byte) (int, error) {
	args := r.Called(p)
	return args.Int(0), args.Error(1)
}

type MockedResponeWriter struct {
	mock.Mock
	httptest.ResponseRecorder
}

func (r *MockedResponeWriter) Write(p []byte) (int, error) {
	args := r.Called(p)
	return args.Int(0), args.Error(1)
}

type MockedLogger struct {
	logging.LoggerStd
	mock.Mock
}

func (l *MockedLogger) Error(a ...interface{}) {
	l.Called(a...)
}

func TestSerializeRequest(t *testing.T) {
	expectedError := errors.New("Read error")
	body := &MockedReader{}
	body.On("Read", mock.Anything).Return(0, expectedError)
	req, err := SerializeRequest(&http.Request{
		Body: body,
	})
	assert.Nil(t, req)
	assert.Equal(t, expectedError, err)
}

func TestSerializeResponse(t *testing.T) {
	expectedError := errors.New("Read error")
	body := &MockedReader{}
	body.On("Read", mock.Anything).Return(0, expectedError)
	res, err := SerializeResponse(&http.Response{
		Body: body,
	})
	assert.Nil(t, res)
	assert.Equal(t, expectedError, err)
}

func TestWriteError(t *testing.T) {
	expectedError := errors.New("Write error")
	logger := &MockedLogger{}
	logger.On("Error", mock.Anything).Return()
	responseWriter := &MockedResponeWriter{}
	responseWriter.On("Write", mock.Anything).Return(0, expectedError)
	WriteError(logger, responseWriter, 500, "Generic error")
	logger.AssertCalled(t, "Error", expectedError)
}

func TestAddrToHost(t *testing.T) {
	assert.Equal(t, "localhost", AddrToHost("localhost"))
	assert.Equal(t, "127.0.0.1", AddrToHost("127.0.0.1:8000"))
}
