package repository

import (
	"acc-server-manager/local/model"
	"context"

	"gorm.io/gorm"
)

type StateHistoryRepository struct {
	db *gorm.DB
}

func NewStateHistoryRepository(db *gorm.DB) *StateHistoryRepository {
	return &StateHistoryRepository{
		db: db,
	}
}

// GetAll
// Gets All rows from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ServerModel: Server object from database.
func (as StateHistoryRepository) GetAll(ctx context.Context, id int) *[]model.StateHistory {
	db := as.db.WithContext(ctx)
	ServerModel := new([]model.StateHistory)
	db.Find(&ServerModel).Where("ID = ?", id)
	return ServerModel
}

// UpdateServer
// Updates Server row from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.Server: Server object from database.
func (as StateHistoryRepository) Insert(ctx context.Context, body *model.StateHistory) *model.StateHistory {
	db := as.db.WithContext(ctx)
	db.Save(body)
	return body
}
