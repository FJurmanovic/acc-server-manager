package repository

import (
	"acc-server-manager/local/model"
	"context"

	"gorm.io/gorm"
)

type ConfigRepository struct {
	*BaseRepository[model.Config, model.ConfigFilter]
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{
		BaseRepository: NewBaseRepository[model.Config, model.ConfigFilter](db, model.Config{}),
	}
}

// UpdateConfig updates or creates a Config record
func (r *ConfigRepository) UpdateConfig(ctx context.Context, config *model.Config) *model.Config {
	if err := r.Update(ctx, config); err != nil {
		// If update fails, try to insert
		if err := r.Insert(ctx, config); err != nil {
			return nil
		}
	}
	return config
}