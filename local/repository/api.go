package repository

import (
	"acc-server-manager/local/model"

	"gorm.io/gorm"
)

type ApiRepository struct {
	*BaseRepository[model.ApiModel, model.ApiFilter]
}

func NewApiRepository(db *gorm.DB) *ApiRepository {
	return &ApiRepository{
		BaseRepository: NewBaseRepository[model.ApiModel, model.ApiFilter](db, model.ApiModel{}),
	}
}
