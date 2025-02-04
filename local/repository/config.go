package repository

import (
	"acc-server-manager/local/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{
		db: db,
	}
}

// UpdateConfig
// Updates first row from Config table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ConfigModel: Config object from database.
func (as ConfigRepository) UpdateFirst(ctx context.Context) *model.Config {
	db := as.db.WithContext(ctx)
	ConfigModel := new(model.Config)
	db.First(&ConfigModel)
	return ConfigModel
}

// UpdateAll
// Updates All rows from Config table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ConfigModel: Config object from database.
func (as ConfigRepository) UpdateAll(ctx context.Context) *[]model.Config {
	db := as.db.WithContext(ctx)
	ConfigModel := new([]model.Config)
	db.Find(&ConfigModel)
	return ConfigModel
}

// UpdateConfig
// Updates Config row from Config table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ConfigModel: Config object from database.
func (as ConfigRepository) UpdateConfig(ctx context.Context, body *model.Config) *model.Config {
	db := as.db.WithContext(ctx)

	existingConfig := new(model.Config)
	result := db.Where("server_id=?", body.ServerID).Where("config_file=?", body.ConfigFile).First(existingConfig)
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		body.ID = existingConfig.ID
	}
	db.Save(body)
	return body
}
