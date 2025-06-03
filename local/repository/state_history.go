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

// GetLastSessionID gets the last session ID for a server
func (r *StateHistoryRepository) GetLastSessionID(ctx context.Context, serverID uint) (uint, error) {
	var lastSession model.StateHistory
	result := r.BaseRepository.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("session_id DESC").
		First(&lastSession)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil // Return 0 if no sessions found
		}
		return 0, result.Error
	}

	return lastSession.SessionID, nil
}
