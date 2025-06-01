package repository

import (
	"acc-server-manager/local/model"
	"context"

	"gorm.io/gorm"
)

type StateHistoryRepository struct {
	*BaseRepository[model.StateHistory, model.StateHistoryFilter]
}

func NewStateHistoryRepository(db *gorm.DB) *StateHistoryRepository {
	return &StateHistoryRepository{
		BaseRepository: NewBaseRepository[model.StateHistory, model.StateHistoryFilter](db, model.StateHistory{}),
	}
}

// GetAll retrieves all state history records with the given filter
func (r *StateHistoryRepository) GetAll(ctx context.Context, filter *model.StateHistoryFilter) (*[]model.StateHistory, error) {
	return r.BaseRepository.GetAll(ctx, filter)
}

// Insert creates a new state history record
func (r *StateHistoryRepository) Insert(ctx context.Context, model *model.StateHistory) error {
	return r.BaseRepository.Insert(ctx, model)
}
