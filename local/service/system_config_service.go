package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"
	"context"
	"fmt"
	"path/filepath"

	"go.uber.org/dig"
)

type SystemConfigService struct {
	repository *repository.SystemConfigRepository
	cache      *model.LookupCache
}

// SystemConfigServiceParams holds the dependencies for SystemConfigService
type SystemConfigServiceParams struct {
	dig.In

	Repository *repository.SystemConfigRepository
	Cache      *model.LookupCache
}

// NewSystemConfigService creates a new SystemConfigService with dependencies injected by dig
func NewSystemConfigService(params SystemConfigServiceParams) *SystemConfigService {
	logging.Debug("Initializing SystemConfigService")
	return &SystemConfigService{
		repository: params.Repository,
		cache:      params.Cache,
	}
}

func (s *SystemConfigService) Initialize(ctx context.Context) error {
	logging.Debug("Initializing system config cache")
	// Cache all configs
	configs, err := s.repository.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get configs for caching: %v", err)
	}

	for _, config := range *configs {
		cacheKey := fmt.Sprintf(model.CacheKeySystemConfig, config.Key)
		s.cache.Set(cacheKey, &config)
		logging.Debug("Cached system config: %s", config.Key)
	}

	logging.Debug("Completed initializing system config cache")
	return nil
}

func (s *SystemConfigService) GetConfig(ctx context.Context, key string) (*model.SystemConfig, error) {
	cacheKey := fmt.Sprintf(model.CacheKeySystemConfig, key)
	
	// Try to get from cache first
	if cached, exists := s.cache.Get(cacheKey); exists {
		if config, ok := cached.(*model.SystemConfig); ok {
			return config, nil
		}
		logging.Debug("Invalid type in cache for key: %s", key)
	}

	// If not in cache, get from database
	logging.Debug("Loading system config from database: %s", key)
	config, err := s.repository.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if config == nil {
		logging.Error("Configuration not found for key: %s", key)
		return nil, nil
	}

	// Cache the result
	s.cache.Set(cacheKey, config)
	return config, nil
}

func (s *SystemConfigService) GetAllConfigs(ctx context.Context) (*[]model.SystemConfig, error) {
	logging.Debug("Loading all system configs from database")
	return s.repository.GetAll(ctx)
}

func (s *SystemConfigService) UpdateConfig(ctx context.Context, config *model.SystemConfig) error {
	if err := s.repository.Update(ctx, config); err != nil {
		return err
	}

	// Update cache
	cacheKey := fmt.Sprintf(model.CacheKeySystemConfig, config.Key)
	s.cache.Set(cacheKey, config)
	logging.Debug("Updated system config in cache: %s", config.Key)
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