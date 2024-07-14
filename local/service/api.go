package service

import (
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/configs"
	"errors"

	"github.com/gofiber/fiber/v2"
)

type ApiService struct {
	Repository *repository.ApiRepository
}

func NewApiService(repository *repository.ApiRepository) *ApiService {
	return &ApiService{
		Repository: repository,
	}
}

/*
GetFirst

Gets first row from API table.

	   	Args:
	   		context.Context: Application context
		Returns:
			string: Application version
*/
func (as ApiService) GetFirst(ctx *fiber.Ctx) string {
	return configs.Version
}

func (as ApiService) ApiStartServer(ctx *fiber.Ctx) (string, error) {
	service, ok := ctx.Locals("service").(string)
	if !ok {
		return "", errors.New("service name missing")
	}
	return as.StartServer(service)
}

func (as ApiService) StartServer(serviceName string) (string, error) {
	return as.ManageService("start", serviceName)
}

func (as ApiService) ApiStopServer(ctx *fiber.Ctx) (string, error) {
	service, ok := ctx.Locals("service").(string)
	if !ok {
		return "", errors.New("service name missing")
	}
	return as.StopServer(service)
}

func (as ApiService) StopServer(serviceName string) (string, error) {
	return as.ManageService("stop", serviceName)
}

func (as ApiService) ApiRestartServer(ctx *fiber.Ctx) (string, error) {
	service, ok := ctx.Locals("service").(string)
	if !ok {
		return "", errors.New("service name missing")
	}
	return as.RestartServer(service)
}

func (as ApiService) RestartServer(serviceName string) (string, error) {
	_, err := as.ManageService("stop", serviceName)
	if err != nil {
		return "", err
	}
	return as.ManageService("start", serviceName)
}

func (as ApiService) ManageService(action string, serviceName string) (string, error) {
	output, err := common.RunElevatedCommand(action, serviceName)
	if err != nil {
		return "", err
	}
	return output, nil
}
