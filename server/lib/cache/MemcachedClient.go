package cache

import (
	"gbf-proxy/lib/marshaler"

	"github.com/bradfitz/gomemcache/memcache"
)

const DEFAULT_MEMCACHED_EXPIRATION = 3600

type MemcachedClient struct {
	*memcache.Client
	marshaler.Marshaler
}

var _ Client = (*MemcachedClient)(nil)

func NewMemcachedClient(mc *memcache.Client, m marshaler.Marshaler) *MemcachedClient {
	return &MemcachedClient{
		Client:    mc,
		Marshaler: m,
	}
}

func (c *MemcachedClient) Get(key string, value interface{}) error {
	item, err := c.Client.Get(key)
	if err != nil {
		return err
	}
	return c.Marshaler.Unmarshal(item.Value, value)
}

func (c *MemcachedClient) Set(key string, value interface{}) error {
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

func (c *MemcachedClient) Has(key string) (bool, error) {
	_, err := c.Client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
