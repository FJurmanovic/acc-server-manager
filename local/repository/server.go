package repository

import (
	"acc-server-manager/local/model"
	"context"
	"errors"

	"gorm.io/gorm"
)

type ServerRepository struct {
	db *gorm.DB
}

func NewServerRepository(db *gorm.DB) *ServerRepository {
	return &ServerRepository{
		db: db,
	}
}

// GetFirst
// Gets first row from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ServerModel: Server object from database.
func (as ServerRepository) GetFirst(ctx context.Context, serverId int) *model.Server {
	db := as.db.WithContext(ctx)
	ServerModel := new(model.Server)
	db.Where("id=?", serverId).First(&ServerModel)
	return ServerModel
}

// GetFirstByServiceName
// Gets first row from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ServerModel: Server object from database.
func (as ServerRepository) GetFirstByServiceName(ctx context.Context, serviceName string) *model.Server {
	db := as.db.WithContext(ctx)
	ServerModel := new(model.Server)
	result := db.Where("service_name=?", serviceName).First(&ServerModel)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return ServerModel
}

// GetAll
// Gets All rows from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			model.ServerModel: Server object from database.
func (as ServerRepository) GetAll(ctx context.Context) *[]model.Server {
	db := as.db.WithContext(ctx)
	ServerModel := new([]model.Server)
	db.Find(&ServerModel)
	return ServerModel
}
