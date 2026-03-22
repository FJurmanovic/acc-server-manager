package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/platform"
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
	activityLog      *ActivityLogService
	statusCache      *model.ServerStatusCache
	runtime          platform.ServerRuntime
}

func NewServiceControlService(
	repository *repository.ServiceControlRepository,
	serverRepository *repository.ServerRepository,
	runtime platform.ServerRuntime,
	activityLog *ActivityLogService,
) *ServiceControlService {
	return &ServiceControlService{
		repository:       repository,
		serverRepository: serverRepository,
		runtime:          runtime,
		activityLog:      activityLog,
		statusCache: model.NewServerStatusCache(model.CacheConfig{
			ExpirationTime: 30 * time.Second,
			ThrottleTime:   5 * time.Second,
			DefaultStatus:  model.StatusRunning,
		}),
	}
}

// getServerFromCtx resolves the server from ctx locals (serverId or service name).
func (as *ServiceControlService) getServerFromCtx(ctx *fiber.Ctx) (*model.Server, error) {
	serviceName, ok := ctx.Locals("service").(string)
	if !ok || serviceName == "" {
		serverId, ok2 := ctx.Locals("serverId").(string)
		if !ok2 || serverId == "" {
			return nil, errors.New("service name missing")
		}
		return as.serverRepository.GetByID(ctx.UserContext(), serverId)
	}
	return as.serverRepository.GetFirstByServiceName(ctx.UserContext(), serviceName)
}

func extractActor(ctx *fiber.Ctx) (actorID, actorUsername string) {
	actorID, _ = ctx.Locals("userID").(string)
	actorUsername, _ = ctx.Locals("actorUsername").(string)
	if actorUsername == "" {
		actorUsername = "system"
	}
	return
}

func (as *ServiceControlService) SetServerService(serverService *ServerService) {
	as.serverService = serverService
}

func (as *ServiceControlService) GetStatus(ctx *fiber.Ctx) (string, error) {
	serviceName, err := as.GetServiceName(ctx)
	if err != nil {
		return "", err
	}

	if status, shouldCheck := as.statusCache.GetStatus(serviceName); !shouldCheck {
		return status.String(), nil
	}

	statusStr, err := as.StatusServer(serviceName)
	if err != nil {
		return "", err
	}

	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)
	return status.String(), nil
}

func (as *ServiceControlService) ServiceControlStartServer(ctx *fiber.Ctx) (string, error) {
	server, err := as.getServerFromCtx(ctx)
	if err != nil {
		return "", err
	}
	serviceName := server.ServiceName

	as.statusCache.UpdateStatus(serviceName, model.StatusStarting)

	_, err = as.StartServer(serviceName)
	if err != nil {
		return "", err
	}
	statusStr, err := as.StatusServer(serviceName)
	if err != nil {
		return "", err
	}

	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)

	if as.activityLog != nil {
		actorID, actorUsername := extractActor(ctx)
		go func() {
			_ = as.activityLog.Log(context.Background(), &model.ActivityLog{
				ServerID: server.ID,
				UserID:   actorID,
				Username: actorUsername,
				Action:   model.ActionServerStart,
				Details:  `{"service":"` + serviceName + `"}`,
			})
		}()
	}

	return status.String(), nil
}

func (as *ServiceControlService) ServiceControlStopServer(ctx *fiber.Ctx) (string, error) {
	server, err := as.getServerFromCtx(ctx)
	if err != nil {
		return "", err
	}
	serviceName := server.ServiceName

	as.statusCache.UpdateStatus(serviceName, model.StatusStopping)

	_, err = as.StopServer(serviceName)
	if err != nil {
		return "", err
	}
	statusStr, err := as.StatusServer(serviceName)
	if err != nil {
		return "", err
	}

	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)

	if as.activityLog != nil {
		actorID, actorUsername := extractActor(ctx)
		go func() {
			_ = as.activityLog.Log(context.Background(), &model.ActivityLog{
				ServerID: server.ID,
				UserID:   actorID,
				Username: actorUsername,
				Action:   model.ActionServerStop,
				Details:  `{"service":"` + serviceName + `"}`,
			})
		}()
	}

	return status.String(), nil
}

func (as *ServiceControlService) ServiceControlRestartServer(ctx *fiber.Ctx) (string, error) {
	server, err := as.getServerFromCtx(ctx)
	if err != nil {
		return "", err
	}
	serviceName := server.ServiceName

	as.statusCache.UpdateStatus(serviceName, model.StatusRestarting)

	_, err = as.RestartServer(serviceName)
	if err != nil {
		return "", err
	}
	statusStr, err := as.StatusServer(serviceName)
	if err != nil {
		return "", err
	}

	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)

	if as.activityLog != nil {
		actorID, actorUsername := extractActor(ctx)
		go func() {
			_ = as.activityLog.Log(context.Background(), &model.ActivityLog{
				ServerID: server.ID,
				UserID:   actorID,
				Username: actorUsername,
				Action:   model.ActionServerRestart,
				Details:  `{"service":"` + serviceName + `"}`,
			})
		}()
	}

	return status.String(), nil
}

func (as *ServiceControlService) StatusServer(serviceName string) (string, error) {
	server, err := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	if err != nil {
		return "", err
	}
	return as.runtime.Status(context.Background(), server.ID)
}

func (as *ServiceControlService) GetCachedStatus(serviceName string) (string, error) {
	if status, shouldCheck := as.statusCache.GetStatus(serviceName); !shouldCheck {
		return status.String(), nil
	}

	statusStr, err := as.StatusServer(serviceName)
	if err != nil {
		return "", err
	}

	status := model.ParseServiceStatus(statusStr)
	as.statusCache.UpdateStatus(serviceName, status)
	return status.String(), nil
}

func (as *ServiceControlService) StartServer(serviceName string) (string, error) {
	server, err := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	if err != nil {
		return "", err
	}

	status, err := as.runtime.Start(context.Background(), server.ID)
	if err != nil {
		return "", err
	}

	as.serverService.StartAccServerRuntime(server)
	return status, err
}

func (as *ServiceControlService) StopServer(serviceName string) (string, error) {
	server, err := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	if err != nil {
		return "", err
	}

	status, err := as.runtime.Stop(context.Background(), server.ID)
	if err != nil {
		return "", err
	}

	as.serverService.instances.Delete(server.ID)
	return status, err
}

func (as *ServiceControlService) RestartServer(serviceName string) (string, error) {
	server, err := as.serverRepository.GetFirstByServiceName(context.Background(), serviceName)
	if err != nil {
		return "", err
	}

	status, err := as.runtime.Restart(context.Background(), server.ID)
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
