package repository

import (
	"acc-server-manager/local/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type ServerRepository struct {
	*BaseRepository[model.Server, model.ServerFilter]
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	repo := &ServerRepository{
		BaseRepository: NewBaseRepository[model.Server, model.ServerFilter](db, model.Server{}),
	}

	return repo
}

// GetFirstByServiceName
// Gets first row from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ServerModel: Server object from database.
func (r *ServerRepository) GetFirstByServiceName(ctx context.Context, serviceName string) (*model.Server, error) {
	result := new(model.Server)
	if err := r.db.WithContext(ctx).Where("service_name = ?", serviceName).First(result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return result, nil
}
