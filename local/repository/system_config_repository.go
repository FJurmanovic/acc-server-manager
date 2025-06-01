package repository

import (
	"acc-server-manager/local/model"
	"context"
	"time"

	"gorm.io/gorm"
)

type SystemConfigRepository struct {
	db *gorm.DB
}

func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{
		db: db,
	}
}

func (r *SystemConfigRepository) Initialize(ctx context.Context) error {
	// Migration and seeding are now handled in the db package
	return nil
}

func (r *SystemConfigRepository) Get(ctx context.Context, key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.db.Where("key = ?", key).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *SystemConfigRepository) GetAll(ctx context.Context) (*[]model.SystemConfig, error) {
	var configs []model.SystemConfig
	if err := r.db.Find(&configs).Error; err != nil {
		return nil, err
	}
	return &configs, nil
}

func (r *SystemConfigRepository) Update(ctx context.Context, config *model.SystemConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	config.DateModified = time.Now().UTC().Format(time.RFC3339)
	return r.db.Model(&model.SystemConfig{}).
		Where("key = ?", config.Key).
		Updates(map[string]interface{}{
			"value":         config.Value,
			"date_modified": config.DateModified,
		}).Error
} 