package applications

import (
	"gbf-proxy/lib/cache"
	"gbf-proxy/lib/hash"
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
	memcached := cache.NewMemcached(memcachedClient, msgpackMarshaler)
	diskCache := cache.NewDiskCache(&hash.Sha1HashFactory{}, msgpackMarshaler)
	caches := []cache.Client{
		memcached,
		diskCache,
	}

	proxyHandler := handlers.NewProxyHandler()
	diskCacheHandler := handlers.NewCacheHandler(proxyHandler, diskCache)
	memcachedHandler := handlers.NewCacheHandler(diskCacheHandler, memcached)
	webHandler := handlers.NewWebHandler(a.Version, a.WebHost, a.WebAddr)
	gatewayHandler := handlers.NewGatewayHandler(a.Version, memcachedHandler, webHandler)
	connectionHandler := handlers.NewConnectionHandler(gatewayHandler)
	service := services.NewListenerService("Proxy", connectionHandler)

	log.Infof("Starting up Granblue Proxy %s", a.Version)
	for _, c := range caches {
		err := c.Start()
		if err != nil {
			return err
		}
	}
	defer cleanupCaches(caches)

	return service.Serve(a.ListenerAddr)
}

func cleanupCaches(caches []cache.Client) {
	for _, c := range caches {
		err := c.Shutdown()
		if err != nil {
			log.Error(err)
		}
	}
}
