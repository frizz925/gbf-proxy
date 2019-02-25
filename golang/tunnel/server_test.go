package tunnel

import (
	"net"
	"net/http"
	"net/url"
	"testing"

	"github.com/Frizz925/gbf-proxy/golang/controller"
	"github.com/PuerkitoBio/goquery"
)

func TestTunnel(t *testing.T) {
	l, err := prepareServices()
	if err != nil {
		t.Fatal(err)
	}

	proxyURL := &url.URL{
		Scheme: "http",
		Host:   l.Addr().String(),
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
	res, err := client.Get("http://game.granbluefantasy.jp")
	if err != nil {
		t.Fatal(err)
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

func prepareServices() (net.Listener, error) {
	c := controller.New(&controller.ServerConfig{})
	l, err := c.Open("localhost:0")
	if err != nil {
		return nil, err
	}
	s := New(&ServerConfig{
		TunnelURL: &url.URL{
			Scheme: "ws",
			Host:   l.Addr().String(),
		},
	})
	return s.Open("localhost:0")
}
