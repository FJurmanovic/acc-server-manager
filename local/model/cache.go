package model

import (
	"acc-server-manager/local/utl/logging"
	"sync"
	"time"
)

type StatusCache struct {
	Status    ServiceStatus
	UpdatedAt time.Time
}

type CacheConfig struct {
	ExpirationTime time.Duration
	ThrottleTime   time.Duration
	DefaultStatus  ServiceStatus
}

type ServerStatusCache struct {
	sync.RWMutex
	cache       map[string]*StatusCache
	config      CacheConfig
	lastChecked map[string]time.Time
}

func NewServerStatusCache(config CacheConfig) *ServerStatusCache {
	return &ServerStatusCache{
		cache:       make(map[string]*StatusCache),
		lastChecked: make(map[string]time.Time),
		config:      config,
	}
}

func (c *ServerStatusCache) GetStatus(serviceName string) (ServiceStatus, bool) {
	c.RLock()
	defer c.RUnlock()

	if lastCheck, exists := c.lastChecked[serviceName]; exists {
		if time.Since(lastCheck) < c.config.ThrottleTime {
			if cached, ok := c.cache[serviceName]; ok {
				return cached.Status, false
			}
			return c.config.DefaultStatus, false
		}
	}

	if cached, ok := c.cache[serviceName]; ok {
		if time.Since(cached.UpdatedAt) < c.config.ExpirationTime {
			return cached.Status, false
		}
	}

	return StatusUnknown, true
}

func (c *ServerStatusCache) UpdateStatus(serviceName string, status ServiceStatus) {
	c.Lock()
	defer c.Unlock()

	c.cache[serviceName] = &StatusCache{
		Status:    status,
		UpdatedAt: time.Now(),
	}
	c.lastChecked[serviceName] = time.Now()
}

func (c *ServerStatusCache) InvalidateStatus(serviceName string) {
	c.Lock()
	defer c.Unlock()

	delete(c.cache, serviceName)
	delete(c.lastChecked, serviceName)
}

func (c *ServerStatusCache) Clear() {
	c.Lock()
	defer c.Unlock()

	c.cache = make(map[string]*StatusCache)
	c.lastChecked = make(map[string]time.Time)
}

type LookupCache struct {
	sync.RWMutex
	data map[string]interface{}
}

func NewLookupCache() *LookupCache {
	logging.Debug("Initializing new LookupCache")
	return &LookupCache{
		data: make(map[string]interface{}),
	}
}

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

func (c *LookupCache) Set(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()

	c.data[key] = value
	logging.Debug("Cache SET for key: %s", key)
}

func (c *LookupCache) Clear() {
	c.Lock()
	defer c.Unlock()

	c.data = make(map[string]interface{})
	logging.Debug("Cache CLEARED")
}

type ConfigEntry[T any] struct {
	Data      T
	UpdatedAt time.Time
}

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

func updateConfigInCache[T any](cache map[string]*ConfigEntry[T], serverID string, data T) {
	cache[serverID] = &ConfigEntry[T]{
		Data:      data,
		UpdatedAt: time.Now(),
	}
	logging.Debug("Config cache SET for server ID: %s", serverID)
}

type ServerConfigCache struct {
	sync.RWMutex
	configuration map[string]*ConfigEntry[Configuration]
	assistRules   map[string]*ConfigEntry[AssistRules]
	event         map[string]*ConfigEntry[EventConfig]
	eventRules    map[string]*ConfigEntry[EventRules]
	settings      map[string]*ConfigEntry[ServerSettings]
	config        CacheConfig
}

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

func (c *ServerConfigCache) GetConfiguration(serverID string) (*Configuration, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get configuration from cache for server ID: %s", serverID)
	return getConfigFromCache(c.configuration, serverID, c.config.ExpirationTime)
}

func (c *ServerConfigCache) GetAssistRules(serverID string) (*AssistRules, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get assist rules from cache for server ID: %s", serverID)
	return getConfigFromCache(c.assistRules, serverID, c.config.ExpirationTime)
}

func (c *ServerConfigCache) GetEvent(serverID string) (*EventConfig, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get event config from cache for server ID: %s", serverID)
	return getConfigFromCache(c.event, serverID, c.config.ExpirationTime)
}

func (c *ServerConfigCache) GetEventRules(serverID string) (*EventRules, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get event rules from cache for server ID: %s", serverID)
	return getConfigFromCache(c.eventRules, serverID, c.config.ExpirationTime)
}

func (c *ServerConfigCache) GetSettings(serverID string) (*ServerSettings, bool) {
	c.RLock()
	defer c.RUnlock()
	logging.Debug("Attempting to get settings from cache for server ID: %s", serverID)
	return getConfigFromCache(c.settings, serverID, c.config.ExpirationTime)
}

func (c *ServerConfigCache) UpdateConfiguration(serverID string, config Configuration) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating configuration cache for server ID: %s", serverID)
	updateConfigInCache(c.configuration, serverID, config)
}

func (c *ServerConfigCache) UpdateAssistRules(serverID string, rules AssistRules) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating assist rules cache for server ID: %s", serverID)
	updateConfigInCache(c.assistRules, serverID, rules)
}

func (c *ServerConfigCache) UpdateEvent(serverID string, event EventConfig) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating event config cache for server ID: %s", serverID)
	updateConfigInCache(c.event, serverID, event)
}

func (c *ServerConfigCache) UpdateEventRules(serverID string, rules EventRules) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating event rules cache for server ID: %s", serverID)
	updateConfigInCache(c.eventRules, serverID, rules)
}

func (c *ServerConfigCache) UpdateSettings(serverID string, settings ServerSettings) {
	c.Lock()
	defer c.Unlock()
	logging.Debug("Updating settings cache for server ID: %s", serverID)
	updateConfigInCache(c.settings, serverID, settings)
}

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
