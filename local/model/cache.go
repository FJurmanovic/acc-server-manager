package model

import (
	"acc-server-manager/local/utl/logging"
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
	logging.Debug("Initializing new LookupCache")
	return &LookupCache{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a cached value by key
func (c *LookupCache) Get(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()
	
	value, exists := c.data[key]
	if exists {
		logging.Debug("Cache HIT for key: %s", key)
	} else {
		logging.Debug("Cache MISS for key: %s", key)
	}
	return value, exists
}

// Set stores a value in the cache
func (c *LookupCache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	
	c.data[key] = value
	logging.Debug("Cache SET for key: %s", key)
}

// Clear removes all entries from the cache
func (c *LookupCache) Clear() {
	c.Lock()
	defer c.Unlock()
	
	c.data = make(map[string]interface{})
	logging.Debug("Cache CLEARED")
}

// ConfigEntry represents a cached configuration entry with its update time
type ConfigEntry[T any] struct {
	Data      T
	UpdatedAt time.Time
}

// getConfigFromCache is a generic helper function to retrieve cached configs
func getConfigFromCache[T any](cache map[string]*ConfigEntry[T], serverID string, expirationTime time.Duration) (*T, bool) {
	if entry, ok := cache[serverID]; ok {
		if time.Since(entry.UpdatedAt) < expirationTime {
			logging.Debug("Config cache HIT for server ID: %s", serverID)
			return &entry.Data, true
		}
		logging.Debug("Config cache EXPIRED for server ID: %s", serverID)
	} else {
		logging.Debug("Config cache MISS for server ID: %s", serverID)
	}
	return nil, false
}

// updateConfigInCache is a generic helper function to update cached configs
func updateConfigInCache[T any](cache map[string]*ConfigEntry[T], serverID string, data T) {
	cache[serverID] = &ConfigEntry[T]{
		Data:      data,
		UpdatedAt: time.Now(),
	}
	logging.Debug("Config cache SET for server ID: %s", serverID)
}

// ServerConfigCache manages cached server configurations
type ServerConfigCache struct {
	sync.RWMutex
	configuration map[string]*ConfigEntry[Configuration]
	assistRules   map[string]*ConfigEntry[AssistRules]
	event         map[string]*ConfigEntry[EventConfig]
	eventRules    map[string]*ConfigEntry[EventRules]
	settings      map[string]*ConfigEntry[ServerSettings]
	config        CacheConfig
}

// NewServerConfigCache creates a new server configuration cache
func NewServerConfigCache(config CacheConfig) *ServerConfigCache {
	logging.Debug("Initializing new ServerConfigCache with expiration time: %v, throttle time: %v", config.ExpirationTime, config.ThrottleTime)
	return &ServerConfigCache{
		configuration: make(map[string]*ConfigEntry[Configuration]),
		assistRules:   make(map[string]*ConfigEntry[AssistRules]),
		event:         make(map[string]*ConfigEntry[EventConfig]),
		eventRules:    make(map[string]*ConfigEntry[EventRules]),
		settings:      make(map[string]*ConfigEntry[ServerSettings]),
		config:        config,
	}
}

// GetConfiguration retrieves a cached configuration
func (c *ServerConfigCache) GetConfiguration(serverID string) (*Configuration, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get configuration from cache for server ID: %s", serverID)
	return getConfigFromCache(c.configuration, serverID, c.config.ExpirationTime)
}

// GetAssistRules retrieves cached assist rules
func (c *ServerConfigCache) GetAssistRules(serverID string) (*AssistRules, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get assist rules from cache for server ID: %s", serverID)
	return getConfigFromCache(c.assistRules, serverID, c.config.ExpirationTime)
}

// GetEvent retrieves cached event configuration
func (c *ServerConfigCache) GetEvent(serverID string) (*EventConfig, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get event config from cache for server ID: %s", serverID)
	return getConfigFromCache(c.event, serverID, c.config.ExpirationTime)
}

// GetEventRules retrieves cached event rules
func (c *ServerConfigCache) GetEventRules(serverID string) (*EventRules, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get event rules from cache for server ID: %s", serverID)
	return getConfigFromCache(c.eventRules, serverID, c.config.ExpirationTime)
}

// GetSettings retrieves cached server settings
func (c *ServerConfigCache) GetSettings(serverID string) (*ServerSettings, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get settings from cache for server ID: %s", serverID)
	return getConfigFromCache(c.settings, serverID, c.config.ExpirationTime)
}

// UpdateConfiguration updates the configuration cache
func (c *ServerConfigCache) UpdateConfiguration(serverID string, config Configuration) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating configuration cache for server ID: %s", serverID)
	updateConfigInCache(c.configuration, serverID, config)
}

// UpdateAssistRules updates the assist rules cache
func (c *ServerConfigCache) UpdateAssistRules(serverID string, rules AssistRules) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating assist rules cache for server ID: %s", serverID)
	updateConfigInCache(c.assistRules, serverID, rules)
}

// UpdateEvent updates the event configuration cache
func (c *ServerConfigCache) UpdateEvent(serverID string, event EventConfig) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating event config cache for server ID: %s", serverID)
	updateConfigInCache(c.event, serverID, event)
}

// UpdateEventRules updates the event rules cache
func (c *ServerConfigCache) UpdateEventRules(serverID string, rules EventRules) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating event rules cache for server ID: %s", serverID)
	updateConfigInCache(c.eventRules, serverID, rules)
}

// UpdateSettings updates the server settings cache
func (c *ServerConfigCache) UpdateSettings(serverID string, settings ServerSettings) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating settings cache for server ID: %s", serverID)
	updateConfigInCache(c.settings, serverID, settings)
}

// InvalidateServerCache removes all cached configurations for a specific server
func (c *ServerConfigCache) InvalidateServerCache(serverID string) {
	c.Lock()
	defer c.Unlock()

	logging.Debug("Invalidating all cache entries for server ID: %s", serverID)
	delete(c.configuration, serverID)
	delete(c.assistRules, serverID)
	delete(c.event, serverID)
	delete(c.eventRules, serverID)
	delete(c.settings, serverID)
}

// Clear removes all entries from the cache
func (c *ServerConfigCache) Clear() {
	c.Lock()
	defer c.Unlock()

	logging.Debug("Clearing all server config cache entries")
	c.configuration = make(map[string]*ConfigEntry[Configuration])
	c.assistRules = make(map[string]*ConfigEntry[AssistRules])
	c.event = make(map[string]*ConfigEntry[EventConfig])
	c.eventRules = make(map[string]*ConfigEntry[EventRules])
	c.settings = make(map[string]*ConfigEntry[ServerSettings])
} 