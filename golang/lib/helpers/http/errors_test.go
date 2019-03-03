package http

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {
	err := NewRequestError(403, "Forbidden", nil)
	expected := "403: Forbidden"
	if err.Error() != expected {
		t.Fatalf("Expects error message to be '%s'", expected)
	}
}

func TestWithBaseError(t *testing.T) {
	base := errors.New("Base error")
	err := NewRequestError(503, "Dummy error", base)
	if err.StatusCode != 503 {
		t.Fatal("Expects error code to be 503")
	}
	if err.Message != "Dummy error" {
		t.Fatal("Expects error code to be 'Dummy error'")
	}
	expected := base.Error()
	if err.Error() != expected {
		t.Fatalf("Expects error message to be '%s'", expected)
	}
}
