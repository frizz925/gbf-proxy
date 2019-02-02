package proxy

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

type TestState struct {
	Listener net.Listener
	URL      string
}

var state *TestState

func TestMain(m *testing.M) {
	s := NewServer()
	l, err := s.Open("localhost:0")
	if err != nil {
		panic(err)
	}
	defer s.Close()
	addr := l.Addr().String()
	url := "http://" + addr
	state = &TestState{
		Listener: l,
		URL:      url,
	}
	os.Exit(m.Run())
}

func TestForbidden(t *testing.T) {
	res, err := http.Get(state.URL)
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
	proxyURL, err := url.Parse(state.URL)
	if err != nil {
		t.Fatal(err)
	}
	c := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	req, err := http.NewRequest("GET", "http://game.granbluefantasy.jp", nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	title := doc.Find("title").Text()
	if title != "グランブルーファンタジー" {
		t.Fatal("Invalid loaded page")
	}
}
