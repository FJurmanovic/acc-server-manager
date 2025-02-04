package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"

	"github.com/gofiber/fiber/v2"
)

type ServerService struct {
	repository *repository.ServerRepository
}

func NewServerService(repository *repository.ServerRepository) *ServerService {
	return &ServerService{
		repository: repository,
	}
}

// GetAll
// Gets All rows from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as ServerService) GetAll(ctx *fiber.Ctx) *[]model.Server {
	return as.repository.GetAll(ctx.UserContext())
}
