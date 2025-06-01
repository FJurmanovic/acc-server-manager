package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ApiService struct {
	repository       *repository.ApiRepository
	serverRepository *repository.ServerRepository
	serverService    *ServerService
	statusCache      *model.ServerStatusCache
	windowsService   *WindowsService
}

func NewApiService(repository *repository.ApiRepository,
	serverRepository *repository.ServerRepository,
	systemConfigService *SystemConfigService) *ApiService {
	return &ApiService{
		repository:       repository,
		serverRepository: serverRepository,
		statusCache: model.NewServerStatusCache(model.CacheConfig{
			ExpirationTime:  30 * time.Second,  // Cache expires after 30 seconds
			ThrottleTime:    5 * time.Second,   // Minimum 5 seconds between checks
			DefaultStatus:   model.StatusRunning, // Default to running if throttled
		}),
		windowsService: NewWindowsService(systemConfigService),
	}
}

func (as *ApiService) SetServerService(serverService *ServerService) {
	as.serverService = serverService
}

func (as *ApiService) GetStatus(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}

	// Try to get status from cache
	if status, shouldCheck := as.statusCache.GetStatus(serviceName); !shouldCheck {
		return status.String(), nil
	}

	// If cache miss or expired, check actual status
	statusStr, err := as.StatusServer(serviceName)
	if err != nil {
		return "", err
	}

	// Parse and update cache with new status
	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)
	return status.String(), nil
}

func (as *ApiService) ApiStartServer(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}
	
	// Update status cache for this service before starting
	as.statusCache.UpdateStatus(serviceName, model.StatusStarting)
	
	statusStr, err := as.StartServer(serviceName)
	if err != nil {
		return "", err
	}
	
	// Parse and update cache with new status
	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)
	return status.String(), nil
}

func (as *ApiService) ApiStopServer(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}
	
	// Update status cache for this service before stopping
	as.statusCache.UpdateStatus(serviceName, model.StatusStopping)
	
	statusStr, err := as.StopServer(serviceName)
	if err != nil {
		return "", err
	}
	
	// Parse and update cache with new status
	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)
	return status.String(), nil
}

func (as *ApiService) ApiRestartServer(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}
	
	// Update status cache for this service before restarting
	as.statusCache.UpdateStatus(serviceName, model.StatusRestarting)
	
	statusStr, err := as.RestartServer(serviceName)
	if err != nil {
		return "", err
	}
	
	// Parse and update cache with new status
	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)
	return status.String(), nil
}

func (as *ApiService) StatusServer(serviceName string) (string, error) {
	return as.windowsService.Status(context.Background(), serviceName)
}

func (as *ApiService) StartServer(serviceName string) (string, error) {
	status, err := as.windowsService.Start(context.Background(), serviceName)
	if err != nil {
		return "", err
	}

	server, err := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	if err != nil {
		return "", err
	}
	as.serverService.StartAccServerRuntime(server)
	return status, err
}

func (as *ApiService) StopServer(serviceName string) (string, error) {
	status, err := as.windowsService.Stop(context.Background(), serviceName)
	if err != nil {
		return "", err
	}

	server, err := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	if err != nil {
		return "", err
	}
	as.serverService.instances.Delete(server.ID)

	return status, err
}

func (as *ApiService) RestartServer(serviceName string) (string, error) {
	status, err := as.windowsService.Restart(context.Background(), serviceName)
	if err != nil {
		return "", err
	}

	server, err := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	if err != nil {
		return "", err
	}
	as.serverService.StartAccServerRuntime(server)
	return status, err
}

func (as *ApiService) GetServiceName(ctx *fiber.Ctx) (string, error) {
	var server *model.Server
	var err error
	serviceName, ok := ctx.Locals("service").(string)
	if !ok || serviceName == "" {
		serverId, ok2 := ctx.Locals("serverId").(int)
		if !ok2 || serverId == 0 {
			return "", errors.New("service name missing")
		}
		server, err = as.serverRepository.GetByID(ctx.UserContext(), serverId)
	} else {
		server, err = as.serverRepository.GetFirstByServiceName(ctx.UserContext(), serviceName)
	}
	if err != nil {
		return "", err
	}
	return server.ServiceName, nil
}
