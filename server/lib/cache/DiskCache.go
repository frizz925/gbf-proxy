package cache

import (
	"fmt"
	"gbf-proxy/lib/hash"
	"gbf-proxy/lib/logger"
	"gbf-proxy/lib/marshaler"
	"io/ioutil"
	"os"
)

type DiskCache struct {
	hash.HashFactory
	marshaler.Marshaler
	*logger.Logger
}

const CACHE_DIRECTORY = "cache"

var _ Client = (*DiskCache)(nil)
var log = logger.DefaultLogger

func NewDiskCache(h hash.HashFactory, m marshaler.Marshaler) *DiskCache {
	return &DiskCache{
		HashFactory: h,
		Marshaler:   m,
		Logger:      logger.DefaultLogger,
	}
}

func (c *DiskCache) Name() string {
	return "Disk"
}

func (c *DiskCache) Start() error {
	c.Logger.Info("Starting up disk cache")
	cacheDir := CACHE_DIRECTORY
	stat, err := os.Stat(cacheDir)
	if os.IsNotExist(err) {
		err := os.Mkdir(cacheDir, os.ModePerm)
		if err != nil {
			return err
		}
	} else if stat != nil {
		if !stat.IsDir() {
			return fmt.Errorf("%s exists and is not a directory", cacheDir)
		}
	} else {
		return err
	}
	return nil
}

func (c *DiskCache) Shutdown() error {
	c.Logger.Info("Shutting down disk cache")
	return nil
}

func (c *DiskCache) Get(key string, value interface{}) error {
	path, err := c.pathKey(key)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	err = c.Marshaler.Unmarshal(b, value)
	return err
}

func (c *DiskCache) Set(key string, value interface{}) error {
	path, err := c.pathKey(key)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	b, err := c.Marshaler.Marshal(value)
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	return err
}

func (c *DiskCache) Has(key string) (bool, error) {
	path, err := c.pathKey(key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (c *DiskCache) pathKey(key string) (string, error) {
	hash, err := c.hashKey(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", CACHE_DIRECTORY, hash), nil
}

func (c *DiskCache) hashKey(key string) (string, error) {
	h := c.HashFactory.New()
	_, err := h.Write([]byte(key))
	if err != nil {
		return "", err
	}
	b := h.Sum(nil)
	return fmt.Sprintf("%x", b), nil
}
