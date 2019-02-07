package cache

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

type testState struct {
	server   *Server
	config   *ServerConfig
	redis    *redis.Client
	client   *http.Client
	listener net.Listener
}

var state *testState

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

func testMainWrapper(m *testing.M) int {
	redis := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer redis.FlushAll()
	err := redis.FlushAll().Err()
	if err != nil {
		panic(err)
	}

	config := &ServerConfig{
		Redis: redis,
	}
	s := New(config)
	l, err := s.Open("localhost:0")
	if err != nil {
		panic(err)
	}

	addr := l.Addr().String()
	proxyURL, err := url.Parse("http://" + addr)
	if err != nil {
		panic(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	state = &testState{
		server:   s.(*Server),
		config:   config,
		redis:    redis,
		client:   client,
		listener: l,
	}
	return m.Run()
}

func TestHeartbeat(t *testing.T) {
	host := state.listener.Addr().String()
	header := make(http.Header)
	header.Set(CacheAPIHeaderName, "1")
	res, err := sendRequest(&http.Request{
		URL: &url.URL{
			Scheme: "http",
			Host:   host,
			Path:   "/ping",
		},
		Header: header,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if text != "OK" {
		t.Fatalf("Response mismatch. Expected: OK, Got: %s", text)
	}
}

func TestCache(t *testing.T) {
	req := &http.Request{
		URL: &url.URL{
			Scheme: "http",
			Host:   "httpbin.org:80",
			Path:   "/json",
		},
	}

	firstResponse, err := sendRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	defer firstResponse.Body.Close()
	firstBody, err := ioutil.ReadAll(firstResponse.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(firstBody) <= 0 {
		t.Fatal("Response body is empty!")
	}

	// HACK: Sleep for a second before sending another request
	time.Sleep(time.Second)
	key := GetKeyForRequest(req)
	err = state.redis.Get(key).Err()
	if err != nil {
		t.Fatal(err)
	}

	secondResponse, err := sendRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	defer secondResponse.Body.Close()
	secondBody, err := ioutil.ReadAll(secondResponse.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(secondBody) <= 0 {
		t.Fatal("Response body is empty!")
	}
	if !reflect.DeepEqual(firstBody, secondBody) {
		t.Fatal("Computed and cached responses don't match!")
	}
}

func sendRequest(req *http.Request) (*http.Response, error) {
	res, err := state.client.Do(req)
	if err != nil {
		return res, err
	}
	code := res.StatusCode
	if code != 200 {
		return res, fmt.Errorf("Request error. Status code: %d", code)
	}
	return res, nil
}
