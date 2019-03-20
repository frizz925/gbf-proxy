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

	"github.com/stretchr/testify/require"
)

type testState struct {
	server   *Server
	config   *ServerConfig
	client   *http.Client
	listener net.Listener
}

var state *testState

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

func testMainWrapper(m *testing.M) int {
	config := &ServerConfig{}
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
	require.Nil(t, err)
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	require.Nil(t, err)
	text := string(b)
	require.Equalf(t, "OK", text, "Response mismatch. Expected: OK, Got: %s", text)
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
	require.Nil(t, err)
	defer firstResponse.Body.Close()
	firstBody, err := ioutil.ReadAll(firstResponse.Body)
	require.Nil(t, err)
	require.NotZero(t, len(firstBody), "Response body is empty!")

	// HACK: Sleep for a second before sending another request
	time.Sleep(time.Second)
	key := GetKeyForRequest(req)
	_, err = state.server.FetchRawFromCache(key)
	require.Nil(t, err)

	secondResponse, err := sendRequest(req)
	require.Nil(t, err)
	defer secondResponse.Body.Close()
	secondBody, err := ioutil.ReadAll(secondResponse.Body)
	require.Nil(t, err)
	require.NotZero(t, len(firstBody), "Response body is empty!")

	require.True(t, reflect.DeepEqual(firstBody, secondBody), "Computed and cached responses don't match!")
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
