package cache

import (
	"gbf-proxy/lib/logger"
	"gbf-proxy/lib/marshaler"

	"github.com/bradfitz/gomemcache/memcache"
)

const DEFAULT_MEMCACHED_EXPIRATION = 86400

type Memcached struct {
	*memcache.Client
	marshaler.Marshaler
	*logger.Logger
}

var _ Client = (*Memcached)(nil)

func NewMemcached(mc *memcache.Client, m marshaler.Marshaler) *Memcached {
	return &Memcached{
		Client:    mc,
		Marshaler: m,
		Logger:    logger.DefaultLogger,
	}
}

func (c *Memcached) Name() string {
	return "Memcached"
}

func (c *Memcached) Start() error {
	c.Logger.Info("Starting up memcached client")
	return nil
}

func (c *Memcached) Shutdown() error {
	c.Logger.Info("Shutting down memcached client")
	return nil
}

func (c *Memcached) Get(key string, value interface{}) error {
	item, err := c.Client.Get(key)
	if err != nil {
		return err
	}
	return c.Marshaler.Unmarshal(item.Value, value)
}

func (c *Memcached) Set(key string, value interface{}) error {
	b, err := c.Marshaler.Marshal(value)
	if err != nil {
		return err
	}
	return c.Client.Set(&memcache.Item{
		Key:        key,
		Value:      b,
		Expiration: DEFAULT_MEMCACHED_EXPIRATION,
	})
}

func (c *Memcached) Has(key string) (bool, error) {
	_, err := c.Client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
