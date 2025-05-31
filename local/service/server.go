package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/tracking"
	"context"
	"path/filepath"
	"strconv"
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
	sessionIDs       sync.Map // Track current session ID per server
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
		logging.Error("Failed to get servers: %v", err)
		return service
	}

	for i := range *servers {
		// Initialize instance regardless of status
		logging.Info("Starting server runtime for server ID: %d", (*servers)[i].ID)
		service.StartAccServerRuntime(&(*servers)[i])
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

func (s *ServerService) getNextSessionID(serverID uint) uint {
	currentID, _ := s.sessionIDs.LoadOrStore(serverID, uint(0))
	nextID := currentID.(uint) + 1
	s.sessionIDs.Store(serverID, nextID)
	return nextID
}

func (s *ServerService) insertStateHistory(serverID uint, state *model.ServerState) {
	// Get or create session ID when session changes
	currentSessionInterface, exists := s.instances.Load(serverID)
	var sessionID uint
	if !exists {
		sessionID = s.getNextSessionID(serverID)
	} else {
		serverInstance := currentSessionInterface.(*tracking.AccServerInstance)
		if serverInstance.State == nil || serverInstance.State.Session != state.Session {
			sessionID = s.getNextSessionID(serverID)
		} else {
			sessionIDInterface, exists := s.sessionIDs.Load(serverID)
			if !exists {
				sessionID = s.getNextSessionID(serverID)
			} else {
				sessionID = sessionIDInterface.(uint)
			}
		}
	}

	s.stateHistoryRepo.Insert(context.Background(), &model.StateHistory{
		ServerID:    serverID,
		Session:     state.Session,
		Track:       state.Track,
		PlayerCount: state.PlayerCount,
		DateCreated: time.Now().UTC(),
		SessionStart: state.SessionStart,
		SessionDurationMinutes: state.SessionDurationMinutes,
		SessionID:   sessionID,
	})
}

func (s *ServerService) updateSessionDuration(server *model.Server, sessionType string) {
	// Get configs using helper methods
	event, err := s.configService.GetEventConfig(server)
	if err != nil {
		logging.Error("Failed to get event config for server %d: %v", server.ID, err)
		return
	}

	configuration, err := s.configService.GetConfiguration(server)
	if err != nil {
		logging.Error("Failed to get configuration for server %d: %v", server.ID, err)
		return
	}

	if instance, ok := s.instances.Load(server.ID); ok {
		serverInstance := instance.(*tracking.AccServerInstance)
		serverInstance.State.Track = event.Track
		serverInstance.State.MaxConnections = configuration.MaxConnections.ToInt()

		// Check if session type has changed
		if serverInstance.State.Session != sessionType {
			// Get new session ID for the new session
			sessionID := s.getNextSessionID(server.ID)
			s.sessionIDs.Store(server.ID, sessionID)
		}

		if sessionType == "" && len(event.Sessions) > 0 {
			sessionType = event.Sessions[0].SessionType
		}
		for _, session := range event.Sessions {
			if session.SessionType == sessionType {
				serverInstance.State.SessionDurationMinutes = session.SessionDurationMinutes.ToInt()
				serverInstance.State.Session = sessionType
				break
			}
		}
	} else {
		logging.Error("No instance found for server ID: %d", server.ID)
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

	// Invalidate config cache for this server before loading new configs
	serverIDStr := strconv.FormatUint(uint64(server.ID), 10)
	s.configService.configCache.InvalidateServerCache(serverIDStr)

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
func (s *ServerService) GetAll(ctx *fiber.Ctx, filter *model.ServerFilter) (*[]model.Server, error) {
	servers, err := s.repository.GetAll(ctx.UserContext(), filter)
	if err != nil {
		logging.Error("Failed to get servers: %v", err)
		return nil, err
	}

	for i := range *servers {
		server := &(*servers)[i]
		status, err := s.apiService.StatusServer(server.ServiceName)
		if err != nil {
			logging.Error("Failed to get status for server %s: %v", server.ServiceName, err)
		}
		(*servers)[i].Status = model.ParseServiceStatus(status)
		instance, ok := s.instances.Load(server.ID)
		if !ok {
			logging.Warn("No instance found for server ID: %d", server.ID)
		} else {
			serverInstance := instance.(*tracking.AccServerInstance)
			if serverInstance.State != nil {
				(*server).State = *serverInstance.State
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
func (as *ServerService) GetById(ctx *fiber.Ctx, serverID int) (*model.Server, error) {
	server, err := as.repository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		return nil, err
	}
	status, err := as.apiService.StatusServer(server.ServiceName)
	if err != nil {
		logging.Error(err.Error())
	}
	server.Status = model.ParseServiceStatus(status)
	instance, ok := as.instances.Load(server.ID)
	if !ok {
		logging.Error("Unable to retrieve instance for server of ID: %d", server.ID)
	} else {
		serverInstance := instance.(*tracking.AccServerInstance)
		if (serverInstance.State != nil) {
			(*server).State = *serverInstance.State
		}
	}

	return server, nil
}