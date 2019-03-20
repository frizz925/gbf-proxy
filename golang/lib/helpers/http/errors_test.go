package http

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := NewRequestError(403, "Forbidden", nil)
	expected := "403: Forbidden"
	assert.Equalf(t, expected, err.Error(), "Expects error message to be '%s'", expected)
}

func TestWithBaseError(t *testing.T) {
	expectedCode := 503
	expectedMessage := "Dummy error"

	base := errors.New("Base error")
	err := NewRequestError(expectedCode, expectedMessage, base)
	assert.Equalf(t, expectedCode, err.StatusCode, "Expects error code to be %d", expectedCode)
	assert.Equalf(t, expectedMessage, err.Message, "Expects error message to be '%s'", expectedMessage)

	expectedError := base.Error()
	assert.Equalf(t, expectedError, err.Error(), "Expects error message to be '%s'", expectedError)
}
