package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/tracking"
	"context"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type ServerService struct {
	repository *repository.ServerRepository
	apiService *ApiService
	instances sync.Map
	configService *ConfigService
}

func NewServerService(repository *repository.ServerRepository, apiService *ApiService, configService *ConfigService) *ServerService {
	service := &ServerService{
		repository: repository,
		apiService: apiService,
		configService: configService,
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

var leaderboardUpdateRegex = regexp.MustCompile(`Updated leaderboard for (\d+) clients`)
var sessionChangeRegex = regexp.MustCompile(`Session changed: (\w+) -> (\w+)`)

func handleLogLine(instance *model.AccServerInstance) func(string) {
	return func (line string) {
		state := (*instance).State
		now := time.Now()
	
		if strings.Contains(line, "client(s) online") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				countStr := parts[1]
				if count, err := strconv.Atoi(countStr); err == nil {
					state.Lock()
					state.PlayerCount = count
					state.Unlock()
				}
			}
		}
	
		if strings.Contains(line, "Session changed") {
			match := sessionChangeRegex.FindStringSubmatch(line)
			if len(match) == 3 {
				newSession := match[2]
	
				state.Lock()
				state.Session = newSession
				state.SessionStart = now
				state.Unlock()
			}
		}
	
		if strings.Contains(line, "Updated leaderboard for") {
			match := leaderboardUpdateRegex.FindStringSubmatch(line)
			if len(match) == 2 {
				if count, err := strconv.Atoi(match[1]); err == nil {
					state.Lock()
					state.PlayerCount = count
					state.Unlock()
				}
			}
		}
	}
}

func (s *ServerService) StartAccServerRuntime(server *model.Server) {
	s.instances.Delete(server.ID)
    instance := &model.AccServerInstance{
        Model: server,
		State: &model.ServerState{PlayerCount: 0},
    }
	config, _  := DecodeFileName(ConfigurationJson)(server.ConfigPath)
	cfg := config.(model.Configuration)
	event, _  := DecodeFileName(EventJson)(server.ConfigPath)
	evt := event.(model.EventConfig)

	instance.State.MaxConnections = cfg.MaxConnections.ToInt()
	instance.State.Track = evt.Track

	go tracking.TailLogFile(filepath.Join(server.ConfigPath, "\\server\\log\\server.log"), handleLogLine(instance))
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
			serverInstance := instance.(*model.AccServerInstance)
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
		serverInstance := instance.(*model.AccServerInstance)
		if (serverInstance.State != nil) {
			server.State = *serverInstance.State
		}
	}

	return server
}