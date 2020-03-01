package cache

import (
	"gbf-proxy/lib/marshaler"

	"github.com/bradfitz/gomemcache/memcache"
)

const DEFAULT_MEMCACHED_EXPIRATION = 86400

type Memcached struct {
	*memcache.Client
	marshaler.Marshaler
}

var _ Client = (*Memcached)(nil)

func NewMemcached(mc *memcache.Client, m marshaler.Marshaler) *Memcached {
	return &Memcached{
		Client:    mc,
		Marshaler: m,
	}
}

func (c *Memcached) Start() error {
	return nil
}

func (c *Memcached) Shutdown() error {
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
