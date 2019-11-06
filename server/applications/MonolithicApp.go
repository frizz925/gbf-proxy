package applications

import (
	"gbf-proxy/lib/cache"
	"gbf-proxy/lib/logger"
	"gbf-proxy/lib/marshaler"
	"gbf-proxy/services"
	"gbf-proxy/services/handlers"
	"net"
	"os"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

type MonolithicApp struct {
	Hostname      string
	MemcachedAddr string
	ListenerAddr  string
}

var _ Application = (*MonolithicApp)(nil)

var log = logger.DefaultLogger

func (a MonolithicApp) Start() error {
	memcachedClient := memcache.New(a.MemcachedAddr)
	msgpackMarshaler := marshaler.NewMsgpackMarshaler()
	cacheClient := cache.NewMemcachedClient(memcachedClient, msgpackMarshaler)

	proxyHandler := handlers.NewProxyHandler()
	cacheHandler := handlers.NewCacheHandler(proxyHandler, cacheClient)
	webHandler := handlers.NewWebHandler(a.Hostname)
	gatewayHandler := handlers.NewGatewayHandler(cacheHandler, webHandler)
	connectionHandler := handlers.NewConnectionHandler(gatewayHandler)
	service := services.NewListenerService(connectionHandler)

	l, err := a.createListener(a.ListenerAddr)
	if err != nil {
		return err
	}
	defer l.Close()
	log.Infof("Proxy listening at %s", a.ListenerAddr)
	return service.Listen(l)
}

func (a MonolithicApp) createListener(addr string) (net.Listener, error) {
	unixPrefix := "unix://"
	if strings.HasPrefix(addr, unixPrefix) {
		unixAddr := strings.ReplaceAll(addr, unixPrefix, "")
		if s, err := os.Stat(unixAddr); !os.IsNotExist(err) {
			err = os.Remove(s.Name())
			if err != nil {
				return nil, err
			}
		}
		l, err := net.Listen("unix", unixAddr)
		if err != nil {
			return nil, err
		}
		return l, os.Chmod(unixAddr, 0666)
	}
	return net.Listen("tcp4", addr)
}
