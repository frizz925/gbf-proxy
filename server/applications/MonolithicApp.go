package applications

import (
	"gbf-proxy/lib/cache"
	"gbf-proxy/lib/logger"
	"gbf-proxy/lib/marshaler"
	"gbf-proxy/services"
	"gbf-proxy/services/handlers"
	"net"

	"github.com/bradfitz/gomemcache/memcache"
)

type MonolithicApp struct{}

var _ Application = (*MonolithicApp)(nil)

var log = logger.Factory.New()

func (MonolithicApp) Start() error {
	memcachedClient := memcache.New("127.0.0.1:11211")
	msgpackMarshaler := marshaler.NewMsgpackMarshaler()
	cacheClient := cache.NewMemcachedClient(memcachedClient, msgpackMarshaler)

	requestHandler := handlers.NewRequestHandler()
	cacheHandler := handlers.NewCacheHandler(requestHandler, cacheClient)
	proxyHandler := handlers.NewProxyHandler(cacheHandler)
	connectionHandler := handlers.NewConnectionHandler(proxyHandler)
	service := services.NewProxyService(connectionHandler)

	l, err := net.Listen("tcp4", "127.0.0.1:8088")
	if err != nil {
		return err
	}
	log.Info("Proxy listening at 127.0.0.1:8088")
	return service.Listen(l)
}
