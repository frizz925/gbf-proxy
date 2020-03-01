package cache

type Client interface {
	Start() error
	Shutdown() error

	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	Has(key string) (bool, error)
}
