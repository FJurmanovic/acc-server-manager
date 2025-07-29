package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ServiceControlService struct {
	repository       *repository.ServiceControlRepository
	serverRepository *repository.ServerRepository
	serverService    *ServerService
	statusCache      *model.ServerStatusCache
	windowsService   *WindowsService
}

func NewServiceControlService(repository *repository.ServiceControlRepository,
	serverRepository *repository.ServerRepository) *ServiceControlService {
	return &ServiceControlService{
		repository:       repository,
		serverRepository: serverRepository,
		statusCache: model.NewServerStatusCache(model.CacheConfig{
			ExpirationTime: 30 * time.Second,    // Cache expires after 30 seconds
			ThrottleTime:   5 * time.Second,     // Minimum 5 seconds between checks
			DefaultStatus:  model.StatusRunning, // Default to running if throttled
		}),
		windowsService: NewWindowsService(),
	}
}

func (as *ServiceControlService) SetServerService(serverService *ServerService) {
	as.serverService = serverService
}

func (as *ServiceControlService) GetStatus(ctx *fiber.Ctx) (string, error) {
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

func (as *ServiceControlService) ServiceControlStartServer(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}

	// Update status cache for this service before starting
	as.statusCache.UpdateStatus(serviceName, model.StatusStarting)

	_, err = as.StartServer(serviceName)
	if err != nil {
		return "", err
	}
	statusStr, err := as.StatusServer(serviceName)
	if err != nil {
		return "", err
	}

	// Parse and update cache with new status
	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)
	return status.String(), nil
}

func (as *ServiceControlService) ServiceControlStopServer(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}

	// Update status cache for this service before stopping
	as.statusCache.UpdateStatus(serviceName, model.StatusStopping)

	_, err = as.StopServer(serviceName)
	if err != nil {
		return "", err
	}
	statusStr, err := as.StatusServer(serviceName)
	if err != nil {
		return "", err
	}

	// Parse and update cache with new status
	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)
	return status.String(), nil
}

func (as *ServiceControlService) ServiceControlRestartServer(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}

	// Update status cache for this service before restarting
	as.statusCache.UpdateStatus(serviceName, model.StatusRestarting)

	_, err = as.RestartServer(serviceName)
	if err != nil {
		return "", err
	}
	statusStr, err := as.StatusServer(serviceName)
	if err != nil {
		return "", err
	}

	// Parse and update cache with new status
	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)
	return status.String(), nil
}

func (as *ServiceControlService) StatusServer(serviceName string) (string, error) {
	return as.windowsService.Status(context.Background(), serviceName)
}

// GetCachedStatus gets the cached status for a service name without requiring fiber context
func (as *ServiceControlService) GetCachedStatus(serviceName string) (string, error) {
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

func (as *ServiceControlService) StartServer(serviceName string) (string, error) {
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

func (as *ServiceControlService) StopServer(serviceName string) (string, error) {
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

func (as *ServiceControlService) RestartServer(serviceName string) (string, error) {
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

func (as *ServiceControlService) GetServiceName(ctx *fiber.Ctx) (string, error) {
	var server *model.Server
	var err error
	serviceName, ok := ctx.Locals("service").(string)
	if !ok || serviceName == "" {
		serverId, ok2 := ctx.Locals("serverId").(string)
		if !ok2 || serverId == "" {
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
