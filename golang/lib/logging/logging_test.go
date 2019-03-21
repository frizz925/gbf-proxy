package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNullWriter(t *testing.T) {
	count := 256
	data := make([]byte, count)
	write, err := NullWriter.Write(data)
	assert.Equal(t, count, write)
	assert.Nil(t, err)
}

type MockedWriter struct {
	mock.Mock
}

func (w *MockedWriter) Write(p []byte) (int, error) {
	args := w.Called(p)
	return args.Int(0), args.Error(1)
}

func TestLogger(t *testing.T) {
	writer := &MockedWriter{}
	errWriter := &MockedWriter{}
	logger := New(&LoggerConfig{
		Name:      "TestLogger",
		Writer:    writer,
		ErrWriter: errWriter,
	}).(*LoggerStd)

	writer.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)
	errWriter.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	logger.Debugf("%d", 1)
	logger.Infof("%d", 2)
	logger.Warnf("%d", 3)
	logger.Errorf("%d", 4)

	writer.AssertNumberOfCalls(t, "Write", 2)
	errWriter.AssertNumberOfCalls(t, "Write", 2)
}
