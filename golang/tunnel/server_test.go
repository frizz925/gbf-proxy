package tunnel

import (
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/Frizz925/gbf-proxy/golang/controller"
	"github.com/Frizz925/gbf-proxy/golang/lib"
	"github.com/PuerkitoBio/goquery"
)

func TestTunnel(t *testing.T) {
	s, err := prepareServices()
	if err != nil {
		t.Fatal(err)
	}
	l, err := s.Open("localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		s.Close()
		s.WaitGroup().Wait()
	}()

	proxyURL := &url.URL{
		Scheme: "http",
		Host:   l.Addr().String(),
	}
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	// Test concurrency
	wg := &sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go concurrencyTest(t, wg, client)
	}
	wg.Wait()
}

func concurrencyTest(t *testing.T, wg *sync.WaitGroup, client *http.Client) {
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
	wg.Done()
}

func prepareServices() (lib.Server, error) {
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
	return s, nil
}
