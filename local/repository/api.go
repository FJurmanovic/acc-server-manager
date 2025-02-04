package repository

import (
	"acc-server-manager/local/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type ApiRepository struct {
	db *gorm.DB
}

func NewApiRepository(db *gorm.DB) *ApiRepository {
	return &ApiRepository{
		db: db,
	}
}

// GetFirst
// Gets first row from API table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ApiModel: Api object from database.
func (as ApiRepository) GetFirst(ctx context.Context) *model.ApiModel {
	db := as.db.WithContext(ctx)
	apiModel := new(model.ApiModel)
	result := db.First(&apiModel)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return apiModel
}
