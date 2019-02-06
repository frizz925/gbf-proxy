package proxy

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/Frizz925/gbf-proxy/golang/controller"
	"github.com/Frizz925/gbf-proxy/golang/lib"
	"github.com/PuerkitoBio/goquery"
)

type testState struct {
	controller lib.Server
	proxy      lib.Server
	client     *http.Client
}

var state *testState

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

func testMainWrapper(m *testing.M) int {
	c := controller.NewServer(&controller.ServerConfig{})
	prepare(c)
	p := NewServer(&ServerConfig{
		BackendAddr: c.Listener().Addr().String(),
	})
	prepare(p)

	defer func() {
		p.Close()
		c.Close()
	}()

	proxyAddr := p.Listener().Addr().String()
	proxyURL, err := url.Parse("http://" + proxyAddr)
	if err != nil {
		panic(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	state = &testState{
		controller: c,
		proxy:      p,
		client:     client,
	}
	return m.Run()
}

func prepare(s lib.Server) lib.Server {
	_, err := s.Open("localhost:0")
	if err != nil {
		panic(err)
	}
	return s
}

func TestForbidden(t *testing.T) {
	res, err := state.client.Get("http://github.com")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	code := res.StatusCode
	if code != 403 {
		t.Fatalf("Request is not forbidden! Status code: %d", code)
	}
}

func TestAllowed(t *testing.T) {
	res, err := state.client.Get("http://game.granbluefantasy.jp")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	code := res.StatusCode
	if code != 200 {
		t.Fatalf("Request error! Status code: %d", code)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	title := doc.Find("title").Text()
	if title != "グランブルーファンタジー" {
		t.Fatal("Invalid loaded page")
	}
}
