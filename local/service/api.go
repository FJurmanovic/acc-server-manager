package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/common"
	"errors"

	"github.com/gofiber/fiber/v2"
)

type ApiService struct {
	repository       *repository.ApiRepository
	serverRepository *repository.ServerRepository
}

func NewApiService(repository *repository.ApiRepository,
	serverRepository *repository.ServerRepository) *ApiService {
	return &ApiService{
		repository:       repository,
		serverRepository: serverRepository,
	}
}

// GetFirst
// Gets first row from API table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as ApiService) GetFirst(ctx *fiber.Ctx) *model.ApiModel {
	return as.repository.GetFirst(ctx.UserContext())
}

func (as ApiService) GetStatus(ctx *fiber.Ctx) (string, error) {
	return as.StatusServer(ctx)
}

func (as ApiService) ApiStartServer(ctx *fiber.Ctx) (string, error) {
	return as.StartServer(ctx)
}

func (as ApiService) ApiStopServer(ctx *fiber.Ctx) (string, error) {
	return as.StopServer(ctx)
}

func (as ApiService) ApiRestartServer(ctx *fiber.Ctx) (string, error) {
	return as.RestartServer(ctx)
}

func (as ApiService) StatusServer(ctx *fiber.Ctx) (string, error) {
	return as.ManageService(ctx, "status")
}

func (as ApiService) StartServer(ctx *fiber.Ctx) (string, error) {
	return as.ManageService(ctx, "start")
}

func (as ApiService) StopServer(ctx *fiber.Ctx) (string, error) {
	return as.ManageService(ctx, "stop")
}

func (as ApiService) RestartServer(ctx *fiber.Ctx) (string, error) {
	return as.ManageService(ctx, "restart")
}

func (as ApiService) ManageService(ctx *fiber.Ctx, action string) (string, error) {
	var server *model.Server
	serviceName, ok := ctx.Locals("service").(string)
	if !ok || serviceName == "" {
		serverId, ok2 := ctx.Locals("serverId").(int)
		if !ok2 || serverId == 0 {
			return "", errors.New("service name missing")
		}
		server = as.serverRepository.GetFirst(ctx.UserContext(), serverId)
	} else {
		server = as.serverRepository.GetFirstByServiceName(ctx.UserContext(), serviceName)
	}
	if server == nil {
		return "", fiber.NewError(404, "Server not found")
	}

	output, err := common.RunElevatedCommand(action, server.ServiceName)
	if err != nil {
		return "", err
	}
	return output, nil
}
