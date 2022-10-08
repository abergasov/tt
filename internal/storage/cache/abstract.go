package cache

//go:generate mockgen -source=abstract.go -destination=abstract_cache_mock.go -package=cache
type Cacher interface {
	Get(key string) (value []byte, ok bool)
	Contains(key string) bool
	Add(key string, value []byte) (evicted bool)
	Shutdown() error
}
