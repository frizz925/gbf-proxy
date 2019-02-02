package proxy

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestServer(t *testing.T) {
	s := NewServer()
	l, err := s.Start("localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	proxyUrl, err := url.Parse("http://" + addr)
	if err != nil {
		t.Fatal(err)
	}
	client := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}
	res, err := client.Get("http://game.granbluefantasy.jp")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	code := res.StatusCode
	if code != 200 {
		t.Fatalf("HTTP error got status code %d", code)
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
