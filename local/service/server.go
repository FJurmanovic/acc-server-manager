package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"

	"github.com/gofiber/fiber/v2"
)

type ServerService struct {
	repository *repository.ServerRepository
	apiService *ApiService
}

func NewServerService(repository *repository.ServerRepository, apiService *ApiService) *ServerService {
	return &ServerService{
		repository: repository,
		apiService: apiService,
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
	servers := as.repository.GetAll(ctx.UserContext())

	for i, server := range *servers {
		status, _ := as.apiService.StatusServer(server.ServiceName)
		(*servers)[i].Status = model.ServiceStatus(status)
	}

	return servers
}

// GetById
// Gets rows by ID from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as ServerService) GetById(ctx *fiber.Ctx, serverID int) *model.Server {
	server := as.repository.GetFirst(ctx.UserContext(), serverID)
	status, _ := as.apiService.StatusServer(server.ServiceName)
	server.Status = model.ServiceStatus(status)

	return server
}

