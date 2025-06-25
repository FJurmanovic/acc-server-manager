package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/cache"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"
	"path/filepath"
	"time"
)

const (
	configCacheDuration = 24 * time.Hour
)

type SystemConfigService struct {
	repository *repository.SystemConfigRepository
	cache      *cache.InMemoryCache
}

// NewSystemConfigService creates a new SystemConfigService with dependencies injected by dig
func NewSystemConfigService(repository *repository.SystemConfigRepository, cache *cache.InMemoryCache) *SystemConfigService {
	logging.Debug("Initializing SystemConfigService")
	return &SystemConfigService{
		repository: repository,
		cache:      cache,
	}
}

func (s *SystemConfigService) GetConfig(ctx context.Context, key string) (*model.SystemConfig, error) {
	cacheKey := fmt.Sprintf(model.CacheKeySystemConfig, key)

	fetcher := func() (*model.SystemConfig, error) {
		logging.Debug("Loading system config from database: %s", key)
		return s.repository.Get(ctx, key)
	}

	return cache.GetOrSet(s.cache, cacheKey, configCacheDuration, fetcher)
}

func (s *SystemConfigService) GetAllConfigs(ctx context.Context) (*[]model.SystemConfig, error) {
	logging.Debug("Loading all system configs from database")
	return s.repository.GetAll(ctx)
}

func (s *SystemConfigService) UpdateConfig(ctx context.Context, config *model.SystemConfig) error {
	if err := s.repository.Update(ctx, config); err != nil {
		return err
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf(model.CacheKeySystemConfig, config.Key)
	s.cache.Delete(cacheKey)
	logging.Debug("Invalidated system config in cache: %s", config.Key)
	return nil
}

func (s *SystemConfigService) GetSteamCMDDirPath(ctx context.Context) (string, error) {
	steamCMDPath, err := s.GetSteamCMDPath(ctx)
	if err != nil {
		return "", err
	}
	return filepath.Dir(steamCMDPath), nil
}

// Helper methods for common configurations
func (s *SystemConfigService) GetSteamCMDPath(ctx context.Context) (string, error) {
	config, err := s.GetConfig(ctx, model.ConfigKeySteamCMDPath)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", nil
	}
	return config.GetEffectiveValue(), nil
}

func (s *SystemConfigService) GetNSSMPath(ctx context.Context) (string, error) {
	config, err := s.GetConfig(ctx, model.ConfigKeyNSSMPath)
	if err != nil {
		return "", err
	}
	if config == nil {
		return "", nil
	}
	return  config.GetEffectiveValue(), nil
} 