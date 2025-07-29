package cache

import (
	"sync"
	"time"

	"acc-server-manager/local/utl/logging"

	"go.uber.org/dig"
)

// CacheItem represents an item in the cache
type CacheItem struct {
	Value      interface{}
	Expiration int64
}

// InMemoryCache is a thread-safe in-memory cache
type InMemoryCache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

// NewInMemoryCache creates and returns a new InMemoryCache instance
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		items: make(map[string]CacheItem),
	}
}

// Set adds an item to the cache with an expiration duration (in seconds)
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

// Get retrieves an item from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		// Item has expired, but don't delete here to avoid lock upgrade.
		// It will be overwritten on the next Set.
		return nil, false
	}

	return item.Value, true
}

// Delete removes an item from the cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// GetOrSet retrieves an item from the cache. If the item is not found, it
// calls the provided function to get the value, sets it in the cache, and
// returns it.
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

// Start initializes the cache and provides it to the DI container.
func Start(di *dig.Container) {
	cache := NewInMemoryCache()
	err := di.Provide(func() *InMemoryCache {
		return cache
	})
	if err != nil {
		logging.Panic("failed to provide cache")
	}
}
