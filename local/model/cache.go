package model

import (
	"sync"
	"time"
)

// StatusCache represents a cached server status with expiration
type StatusCache struct {
	Status    ServiceStatus
	UpdatedAt time.Time
}

// CacheConfig holds configuration for cache behavior
type CacheConfig struct {
	ExpirationTime  time.Duration // How long before a cache entry expires
	ThrottleTime    time.Duration // Minimum time between status checks
	DefaultStatus   ServiceStatus // Default status to return when throttled
}

// ServerStatusCache manages cached server statuses
type ServerStatusCache struct {
	sync.RWMutex
	cache       map[string]*StatusCache
	config      CacheConfig
	lastChecked map[string]time.Time
}

// NewServerStatusCache creates a new server status cache
func NewServerStatusCache(config CacheConfig) *ServerStatusCache {
	return &ServerStatusCache{
		cache:       make(map[string]*StatusCache),
		lastChecked: make(map[string]time.Time),
		config:      config,
	}
}

// GetStatus retrieves the cached status or indicates if a fresh check is needed
func (c *ServerStatusCache) GetStatus(serviceName string) (ServiceStatus, bool) {
	c.RLock()
	defer c.RUnlock()

	// Check if we're being throttled
	if lastCheck, exists := c.lastChecked[serviceName]; exists {
		if time.Since(lastCheck) < c.config.ThrottleTime {
			if cached, ok := c.cache[serviceName]; ok {
				return cached.Status, false
			}
			return c.config.DefaultStatus, false
		}
	}

	// Check if we have a valid cached entry
	if cached, ok := c.cache[serviceName]; ok {
		if time.Since(cached.UpdatedAt) < c.config.ExpirationTime {
			return cached.Status, false
		}
	}

	return StatusUnknown, true
}

// UpdateStatus updates the cache with a new status
func (c *ServerStatusCache) UpdateStatus(serviceName string, status ServiceStatus) {
	c.Lock()
	defer c.Unlock()

	c.cache[serviceName] = &StatusCache{
		Status:    status,
		UpdatedAt: time.Now(),
	}
	c.lastChecked[serviceName] = time.Now()
}

// Clear removes all entries from the cache
func (c *ServerStatusCache) Clear() {
	c.Lock()
	defer c.Unlock()

	c.cache = make(map[string]*StatusCache)
	c.lastChecked = make(map[string]time.Time)
}

// LookupCache provides a generic cache for lookup data
type LookupCache struct {
	sync.RWMutex
	data map[string]interface{}
}

// NewLookupCache creates a new lookup cache
func NewLookupCache() *LookupCache {
	return &LookupCache{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a cached value by key
func (c *LookupCache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()
	
	value, exists := c.data[key]
	return value, exists
}

// Set stores a value in the cache
func (c *LookupCache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	
	c.data[key] = value
}

// Clear removes all entries from the cache
func (c *LookupCache) Clear() {
	c.Lock()
	defer c.Unlock()
	
	c.data = make(map[string]interface{})
} 