package applications

import (
	"gbf-proxy/lib/cache"
	"gbf-proxy/lib/logger"
	"gbf-proxy/lib/marshaler"
	"gbf-proxy/services"
	"gbf-proxy/services/handlers"

	"github.com/bradfitz/gomemcache/memcache"
)

type MonolithicApp struct {
	Version       string
	WebAddr       string
	WebHost       string
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
	webHandler := handlers.NewWebHandler(a.Version, a.WebHost, a.WebAddr)
	gatewayHandler := handlers.NewGatewayHandler(a.Version, cacheHandler, webHandler)
	connectionHandler := handlers.NewConnectionHandler(gatewayHandler)
	service := services.NewListenerService("Proxy", connectionHandler)

	log.Infof("Starting up Granblue Proxy %s", a.Version)
	return service.Serve(a.ListenerAddr)
}
