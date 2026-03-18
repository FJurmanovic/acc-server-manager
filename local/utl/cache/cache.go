package cache

import (
	"sync"
	"time"

	"acc-server-manager/local/utl/logging"

	"go.uber.org/dig"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type InMemoryCache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		items: make(map[string]CacheItem),
	}
}

func (c *InMemoryCache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func GetOrSet[T any](c *InMemoryCache, key string, duration time.Duration, fetcher func() (T, error)) (T, error) {
	if cached, found := c.Get(key); found {
		if value, ok := cached.(T); ok {
			return value, nil
		}
	}

	value, err := fetcher()
	if err != nil {
		var zero T
		return zero, err
	}

	c.Set(key, value, duration)
	return value, nil
}

func Start(di *dig.Container) {
	cache := NewInMemoryCache()
	err := di.Provide(func() *InMemoryCache {
		return cache
	})
	if err != nil {
		logging.Panic("failed to provide cache")
	}
}
