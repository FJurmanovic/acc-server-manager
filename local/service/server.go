package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/tracking"
	"context"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ServerService struct {
	repository *repository.ServerRepository
	stateHistoryRepo *repository.StateHistoryRepository
	apiService *ApiService
	instances sync.Map
	configService *ConfigService
}

func NewServerService(repository *repository.ServerRepository, stateHistoryRepo *repository.StateHistoryRepository, apiService *ApiService, configService *ConfigService) *ServerService {
	service := &ServerService{
		repository: repository,
		apiService: apiService,
		configService: configService,
		stateHistoryRepo: stateHistoryRepo,
	}
	servers := repository.GetAll(context.Background())
	for _, server := range *servers {
		status, err := service.apiService.StatusServer(server.ServiceName)
		if err != nil {
			log.Print(err.Error())
		}
		if (status == string(model.StatusRunning)) {
			service.StartAccServerRuntime(&server)
		}
	}
	return service
}

func (s *ServerService) StartAccServerRuntime(server *model.Server) {
	s.instances.Delete(server.ID)
    instance := tracking.NewAccServerInstance(server, func(state *model.ServerState, states ...tracking.StateChange) {
		s.stateHistoryRepo.Insert(context.Background(), &model.StateHistory{
			ServerID: server.ID,
			Session: state.Session,
			PlayerCount: state.PlayerCount,
			DateCreated: time.Now().UTC(),
		})
	})
	config, _  := DecodeFileName(ConfigurationJson)(server.ConfigPath)
	cfg := config.(model.Configuration)
	event, _  := DecodeFileName(EventJson)(server.ConfigPath)
	evt := event.(model.EventConfig)

	instance.State.MaxConnections = cfg.MaxConnections.ToInt()
	instance.State.Track = evt.Track

	go tracking.TailLogFile(filepath.Join(server.ConfigPath, "\\server\\log\\server.log"), instance.HandleLogLine)
    s.instances.Store(server.ID, instance)
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
		status, err := as.apiService.StatusServer(server.ServiceName)
		if err != nil {
			log.Print(err.Error())
		}
		(*servers)[i].Status = model.ServiceStatus(status)
		instance, ok := as.instances.Load(server.ID)
		if !ok {
			log.Print("Unable to retrieve instance for server of ID: ", server.ID)
		} else {
			serverInstance := instance.(*tracking.AccServerInstance)
			if (serverInstance.State != nil) {
				(*servers)[i].State = *serverInstance.State
			}
		}
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
	status, err := as.apiService.StatusServer(server.ServiceName)
	if err != nil {
		log.Print(err.Error())
	}
	server.Status = model.ServiceStatus(status)
	instance, ok := as.instances.Load(server.ID)
	if !ok {
		log.Print("Unable to retrieve instance for server of ID: ", server.ID)
	} else {
		serverInstance := instance.(*tracking.AccServerInstance)
		if (serverInstance.State != nil) {
			server.State = *serverInstance.State
		}
	}

	return server
}