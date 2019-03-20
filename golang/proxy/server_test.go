package proxy

import (
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/Frizz925/gbf-proxy/golang/lib"
	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

type testState struct {
	proxy  *Server
	client *http.Client
}

var state *testState

func TestMain(m *testing.M) {
	os.Exit(testMainWrapper(m))
}

func testMainWrapper(m *testing.M) int {
	p := New(&ServerConfig{
		BackendAddr: "game.granbluefantasy.jp:80",
	})
	prepare(p)

	defer func() {
		p.Close()
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
		proxy:  p.(*Server),
		client: client,
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

func TestProxy(t *testing.T) {
	res, err := state.client.Get("http://game.granbluefantasy.jp")
	require.Nil(t, err)
	defer res.Body.Close()
	code := res.StatusCode
	require.Equalf(t, 200, code, "Request error! Status code: %d", code)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	require.Nil(t, err)
	title := doc.Find("title").Text()
	require.Equal(t, "グランブルーファンタジー", title, "Invalid loaded page")
}
