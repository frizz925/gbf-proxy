package controller

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
)

type testWebServer struct {
	content string
}

func TestForbidden(t *testing.T) {
	s := NewServer(&ServerConfig{
		WebAddr: "0.0.0.0:80",
		WebHost: "not-localhost",
	})
	l, err := s.Open("localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	res, err := makeRequest(l)
	if err != nil {
		t.Fatal(err)
	}
	code := res.StatusCode
	if code != 403 {
		t.Fatalf("Request not forbidden! Status code: %d", code)
	}
}

func TestWebServer(t *testing.T) {
	expectedResponse := "Granblue Proxy Web Server"
	w, err := createWebServer(expectedResponse)
	if err != nil {
		t.Fatal(err)
	}
	addr := w.Addr().String()
	config := &ServerConfig{
		WebAddr: addr,
	}
	s := NewServer(config)
	l, err := s.Open("localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	config.WebHost = l.Addr().String()
	res, err := makeRequest(l)
	if err != nil {
		t.Fatal(err)
	}
	code := res.StatusCode
	if code != 200 {
		t.Fatalf("Request error! Got status code %d", code)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	bodyText := string(body)
	if bodyText != expectedResponse {
		t.Fatalf("Response mismatch! Expected: %s, got: %s", expectedResponse, bodyText)
	}
}

func makeRequest(l net.Listener) (*http.Response, error) {
	addr := l.Addr().String()
	return http.Get("http://" + addr)
}

func createWebServer(content string) (net.Listener, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, err
	}
	go func() {
		err := http.Serve(l, &testWebServer{
			content: content,
		})
		if err != nil {
			// do nothing
		}
	}()
	return l, nil
}

func (s *testWebServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	_, err := w.Write([]byte(s.content))
	if err != nil && err != io.EOF {
		panic(err)
	}
}
