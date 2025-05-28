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
	repository       *repository.ServerRepository
	stateHistoryRepo *repository.StateHistoryRepository
	apiService       *ApiService
	instances        sync.Map
	configService    *ConfigService
	lastInsertTimes  sync.Map // Track last insert time per server
	debouncers       sync.Map // Track debounce timers per server
	logTailers       sync.Map // Track log tailers per server
	sessionCache     sync.Map // Cache of server event sessions
}

type pendingState struct {
	timer *time.Timer
	state *model.ServerState
}

func (s *ServerService) ensureLogTailing(server *model.Server, instance *tracking.AccServerInstance) {
	// Check if we already have a tailer
	if _, exists := s.logTailers.Load(server.ID); exists {
		return
	}

	// Start tailing in a goroutine that handles file creation/deletion
	go func() {
		logPath := filepath.Join(server.ConfigPath, "\\server\\log\\server.log")
		tailer := tracking.NewLogTailer(logPath, instance.HandleLogLine)
		s.logTailers.Store(server.ID, tailer)
		
		// Start tailing and automatically handle file changes
		tailer.Start()
	}()
}

func NewServerService(repository *repository.ServerRepository, stateHistoryRepo *repository.StateHistoryRepository, apiService *ApiService, configService *ConfigService) *ServerService {
	service := &ServerService{
		repository:       repository,
		apiService:       apiService,
		configService:    configService,
		stateHistoryRepo: stateHistoryRepo,
	}

	// Initialize instances for all servers
	servers, err := repository.GetAll(context.Background(), &model.ServerFilter{})
	if err != nil {
		log.Print(err.Error())
		return service
	}

	for _, server := range *servers {
		// Initialize instance regardless of status
		service.StartAccServerRuntime(&server)
	}

	return service
}

func (s *ServerService) shouldInsertStateHistory(serverID uint) bool {
	insertInterval := 5 * time.Minute // Configure this as needed
	
	lastInsertInterface, exists := s.lastInsertTimes.Load(serverID)
	if !exists {
		s.lastInsertTimes.Store(serverID, time.Now().UTC())
		return true
	}
	
	lastInsert := lastInsertInterface.(time.Time)
	now := time.Now().UTC()
	
	if now.Sub(lastInsert) >= insertInterval {
		s.lastInsertTimes.Store(serverID, now)
		return true
	}
	
	return false
}

func (s *ServerService) insertStateHistory(serverID uint, state *model.ServerState) {
	s.stateHistoryRepo.Insert(context.Background(), &model.StateHistory{
		ServerID:    serverID,
		Session:     state.Session,
		PlayerCount: state.PlayerCount,
		DateCreated: time.Now().UTC(),
		SessionDurationMinutes: state.SessionDurationMinutes,
	})
}

func (s *ServerService) updateSessionDuration(server *model.Server, sessionType string) {
	sessionsInterface, exists := s.sessionCache.Load(server.ID)
	if !exists {
		// Try to load sessions from config
		event, err := DecodeFileName(EventJson)(server.ConfigPath)
		if err != nil {
			log.Printf("Failed to load event config for server %d: %v", server.ID, err)
			return
		}
		evt := event.(model.EventConfig)
		s.sessionCache.Store(server.ID, evt.Sessions)
		sessionsInterface = evt.Sessions
	}

	sessions := sessionsInterface.([]model.Session)
	if (sessionType == "" && len(sessions) > 0) {
		sessionType = sessions[0].SessionType
	}
	for _, session := range sessions {
		if session.SessionType == sessionType {
			if instance, ok := s.instances.Load(server.ID); ok {
				serverInstance := instance.(*tracking.AccServerInstance)
				serverInstance.State.SessionDurationMinutes = session.SessionDurationMinutes.ToInt()
				serverInstance.State.Session = sessionType
			}
			break
		}
	}
}

func (s *ServerService) handleStateChange(server *model.Server, state *model.ServerState) {
	// Update session duration when session changes
	s.updateSessionDuration(server, state.Session)

	// Cancel existing timer if any
	if debouncer, exists := s.debouncers.Load(server.ID); exists {
		pending := debouncer.(*pendingState)
		pending.timer.Stop()
	}

	// Create new timer
	timer := time.NewTimer(5 * time.Minute)
	s.debouncers.Store(server.ID, &pendingState{
		timer: timer,
		state: state,
	})

	// Start goroutine to handle the delayed insert
	go func() {
		<-timer.C
		if debouncer, exists := s.debouncers.Load(server.ID); exists {
			pending := debouncer.(*pendingState)
			s.insertStateHistory(server.ID, pending.state)
			s.debouncers.Delete(server.ID)
		}
	}()

	// If enough time has passed since last insert, insert immediately
	if s.shouldInsertStateHistory(server.ID) {
		s.insertStateHistory(server.ID, state)
	}
}

func (s *ServerService) StartAccServerRuntime(server *model.Server) {
	// Get or create instance
	instanceInterface, exists := s.instances.Load(server.ID)
	var instance *tracking.AccServerInstance
	if !exists {
		instance = tracking.NewAccServerInstance(server, func(state *model.ServerState, states ...tracking.StateChange) {
			s.handleStateChange(server, state)
		})
		s.instances.Store(server.ID, instance)
	} else {
		instance = instanceInterface.(*tracking.AccServerInstance)
	}

	config, _  := DecodeFileName(ConfigurationJson)(server.ConfigPath)
	cfg := config.(model.Configuration)
	event, _  := DecodeFileName(EventJson)(server.ConfigPath)
	evt := event.(model.EventConfig)

	instance.State.MaxConnections = cfg.MaxConnections.ToInt()
	instance.State.Track = evt.Track

	// Cache sessions for duration lookup
	s.sessionCache.Store(server.ID, evt.Sessions)
	
	s.updateSessionDuration(server, instance.State.Session)

	// Ensure log tailing is running (regardless of server status)
	s.ensureLogTailing(server, instance)
}

// GetAll
// Gets All rows from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as ServerService) GetAll(ctx *fiber.Ctx, filter *model.ServerFilter) (*[]model.Server, error) {
	servers, err := as.repository.GetAll(ctx.UserContext(), filter)
	if err != nil {
		return nil, err
	}

	for i, server := range *servers {
		status, err := as.apiService.StatusServer(server.ServiceName)
		if err != nil {
			log.Print(err.Error())
		}
		(*servers)[i].Status = model.ParseServiceStatus(status)
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

	return servers, nil
}

// GetById
// Gets rows by ID from Server table.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as ServerService) GetById(ctx *fiber.Ctx, serverID int) (*model.Server, error) {
	server, err := as.repository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		return nil, err
	}
	status, err := as.apiService.StatusServer(server.ServiceName)
	if err != nil {
		log.Print(err.Error())
	}
	server.Status = model.ParseServiceStatus(status)
	instance, ok := as.instances.Load(server.ID)
	if !ok {
		log.Print("Unable to retrieve instance for server of ID: ", server.ID)
	} else {
		serverInstance := instance.(*tracking.AccServerInstance)
		if (serverInstance.State != nil) {
			server.State = *serverInstance.State
		}
	}

	return server, nil
}