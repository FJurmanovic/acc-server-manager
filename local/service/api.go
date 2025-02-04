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
	service := ctx.Params("service")
	return as.StatusServer(ctx, service)
}

func (as ApiService) ApiStartServer(ctx *fiber.Ctx) (string, error) {
	service, ok := ctx.Locals("service").(string)
	if !ok {
		return "", errors.New("service name missing")
	}
	return as.StartServer(ctx, service)
}

func (as ApiService) ApiStopServer(ctx *fiber.Ctx) (string, error) {
	service, ok := ctx.Locals("service").(string)
	if !ok {
		return "", errors.New("service name missing")
	}
	return as.StopServer(ctx, service)
}

func (as ApiService) ApiRestartServer(ctx *fiber.Ctx) (string, error) {
	service, ok := ctx.Locals("service").(string)
	if !ok {
		return "", errors.New("service name missing")
	}
	return as.RestartServer(ctx, service)
}

func (as ApiService) StatusServer(ctx *fiber.Ctx, serviceName string) (string, error) {
	return as.ManageService(ctx, "status", serviceName)
}

func (as ApiService) StartServer(ctx *fiber.Ctx, serviceName string) (string, error) {
	return as.ManageService(ctx, "start", serviceName)
}

func (as ApiService) StopServer(ctx *fiber.Ctx, serviceName string) (string, error) {
	return as.ManageService(ctx, "stop", serviceName)
}

func (as ApiService) RestartServer(ctx *fiber.Ctx, serviceName string) (string, error) {
	return as.ManageService(ctx, "restart", serviceName)
}

func (as ApiService) ManageService(ctx *fiber.Ctx, action string, serviceName string) (string, error) {
	server := as.serverRepository.GetFirstByServiceName(ctx.UserContext(), serviceName)
	if server == nil {
		return "", fiber.NewError(404, "Server not found")
	}

	output, err := common.RunElevatedCommand(action, serviceName)
	if err != nil {
		return "", err
	}
	return output, nil
}
