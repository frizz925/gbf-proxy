package http

import "fmt"

type RequestError struct {
	base       error
	StatusCode int
	Message    string
}

func NewRequestError(code int, message string, err error) *RequestError {
	return &RequestError{
		base:       err,
		StatusCode: code,
		Message:    message,
	}
}

func (e *RequestError) Error() string {
	if e.base != nil {
		return e.base.Error()
	}
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}
