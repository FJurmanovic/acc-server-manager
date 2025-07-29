package repository

import (
	"acc-server-manager/local/model"

	"gorm.io/gorm"
)

type ServiceControlRepository struct {
	*BaseRepository[model.ServiceControlModel, model.ServiceControlFilter]
}

func NewServiceControlRepository(db *gorm.DB) *ServiceControlRepository {
	return &ServiceControlRepository{
		BaseRepository: NewBaseRepository[model.ServiceControlModel, model.ServiceControlFilter](db, model.ServiceControlModel{}),
	}
}
