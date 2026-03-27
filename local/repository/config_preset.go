package repository

import (
	"acc-server-manager/local/model"

	"gorm.io/gorm"
)

type ConfigPresetRepository struct {
	*BaseRepository[model.ConfigPreset, model.ConfigPresetFilter]
}

func NewConfigPresetRepository(db *gorm.DB) *ConfigPresetRepository {
	return &ConfigPresetRepository{
		BaseRepository: NewBaseRepository[model.ConfigPreset, model.ConfigPresetFilter](db, model.ConfigPreset{}),
	}
}
