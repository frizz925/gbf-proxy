package controller

import (
	"net/http"
	"testing"
)

func TestForbidden(t *testing.T) {
	s := NewServer(&ServerConfig{})
	l, err := s.Open("localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	res, err := http.Get("http://" + addr)
	if err != nil {
		t.Fatal(err)
	}
	code := res.StatusCode
	if code != 403 {
		t.Fatalf("Request not forbidden! Status code: %d", code)
	}
}
