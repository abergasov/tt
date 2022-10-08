package cache

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"
)

type LRU struct {
	*lru.Cache
}

func NewCache() (Cacher, error) {
	lruCache, err := lru.New(1024)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}
	println("loading cache...")
	println("loading cache done")
	return &LRU{
		lruCache,
	}, nil
}

func (l *LRU) Get(key string) (value []byte, ok bool) {
	data, ok := l.Cache.Get(key)
	if !ok {
		return nil, false
	}
	result, ok := data.([]byte)
	return result, ok
}

func (l *LRU) Contains(key string) bool {
	return l.Cache.Contains(key)
}

func (l *LRU) Add(key string, value []byte) (evicted bool) {
	return l.Cache.Add(key, value)
}

func (l *LRU) Shutdown() error {
	println("dumping cache...")
	println("dumping cache done")
	return nil
}
