package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/common"
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)


type ApiService struct {
	repository       *repository.ApiRepository
	serverRepository *repository.ServerRepository
	serverService *ServerService
}

func NewApiService(repository *repository.ApiRepository,
	serverRepository *repository.ServerRepository,) *ApiService {
	return &ApiService{
		repository:       repository,
		serverRepository: serverRepository,
	}
}

func (as ApiService) SetServerService(serverService *ServerService) {
	as.serverService = serverService
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
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}
	status, err := as.StatusServer(serviceName)

	return status, err
}

func (as ApiService) ApiStartServer(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}
	return as.StartServer(serviceName)
}

func (as ApiService) ApiStopServer(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}
	return as.StopServer(serviceName)
}

func (as ApiService) ApiRestartServer(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}
	return as.RestartServer(serviceName)
}

func (as ApiService) StatusServer(serviceName string) (string, error) {
	return ManageService(serviceName, "status")
}

func (as ApiService) StartServer(serviceName string) (string, error) {
	status, err := ManageService(serviceName, "start")

	server := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	as.serverService.StartAccServerRuntime(server)
	return status, err
}

func (as ApiService) StopServer(serviceName string) (string, error) {
	status, err := ManageService(serviceName, "stop")

	server := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	as.serverService.instances.Delete(server.ID)

	return status, err
}

func (as ApiService) RestartServer(serviceName string) (string, error) {
	status, err := ManageService(serviceName, "restart")

	server := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	as.serverService.StartAccServerRuntime(server)
	return status, err
}

func ManageService(serviceName string, action string) (string, error) {
	output, err := common.RunElevatedCommand(action, serviceName)
	if err != nil {
		return "", err
	}

	return strings.ReplaceAll(output, "\x00", ""), nil
}

func (as ApiService) GetServiceName(ctx *fiber.Ctx) (string, error) {
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
	return server.ServiceName, nil
}
