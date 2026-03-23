package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/platform"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/env"
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/network"
	"acc-server-manager/local/utl/tracking"
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	DefaultStartPort  = 9600
	RequiredPortCount = 1
)

type ServerService struct {
	repository       *repository.ServerRepository
	stateHistoryRepo *repository.StateHistoryRepository
	activityLogRepo  *repository.ActivityLogRepository
	apiService       *ServiceControlService
	configService    *ConfigService
	steamService     *SteamService
	runtime          platform.ServerRuntime
	firewall         platform.FirewallManager
	logStreamer       platform.LogStreamer
	portPool         *network.PortPoolManager
	webSocketService *WebSocketService
	instances        sync.Map // uuid.UUID → *tracking.AccServerInstance
	lastInsertTimes  sync.Map // uuid.UUID → time.Time
	debouncers       sync.Map // uuid.UUID → *pendingState
	sessionIDs       sync.Map // uuid.UUID → uuid.UUID
}

type pendingState struct {
	timer *time.Timer
	state *model.ServerState
}

func (s *ServerService) ensureLogTailing(server *model.Server, instance *tracking.AccServerInstance) {
	go func() {
		if err := s.logStreamer.Start(context.Background(), server, instance.HandleLogLine); err != nil {
			logging.Error("Failed to start log streaming for server %s: %v", server.ID, err)
		}
	}()
}

func NewServerService(
	repository *repository.ServerRepository,
	stateHistoryRepo *repository.StateHistoryRepository,
	activityLogRepo *repository.ActivityLogRepository,
	apiService *ServiceControlService,
	configService *ConfigService,
	steamService *SteamService,
	runtime platform.ServerRuntime,
	firewall platform.FirewallManager,
	logStreamer platform.LogStreamer,
	portPool *network.PortPoolManager,
	webSocketService *WebSocketService,
) *ServerService {
	service := &ServerService{
		repository:       repository,
		stateHistoryRepo: stateHistoryRepo,
		activityLogRepo:  activityLogRepo,
		apiService:       apiService,
		configService:    configService,
		steamService:     steamService,
		runtime:          runtime,
		firewall:         firewall,
		logStreamer:      logStreamer,
		portPool:         portPool,
		webSocketService: webSocketService,
	}

	servers, err := repository.GetAll(context.Background(), &model.ServerFilter{})
	if err != nil {
		logging.Error("Failed to get servers: %v", err)
		return service
	}

	for i := range *servers {
		logging.Info("Starting server runtime for server ID: %d", (*servers)[i].ID)
		service.StartAccServerRuntime(&(*servers)[i])
	}

	return service
}

func (s *ServerService) shouldInsertStateHistory(serverID uuid.UUID) bool {
	insertInterval := 5 * time.Minute

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

func (s *ServerService) getNextSessionID(serverID uuid.UUID) uuid.UUID {
	lastID, err := s.stateHistoryRepo.GetLastSessionID(context.Background(), serverID)
	if err != nil {
		logging.Error("Failed to get last session ID for server %s: %v", serverID, err)
		return uuid.New()
	}
	if lastID == uuid.Nil {
		return uuid.New()
	}
	return uuid.New()
}

func (s *ServerService) insertStateHistory(serverID uuid.UUID, state *model.ServerState) {
	currentSessionInterface, exists := s.instances.Load(serverID)
	var sessionID uuid.UUID
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
				sessionID = sessionIDInterface.(uuid.UUID)
			}
		}
	}

	s.stateHistoryRepo.Insert(context.Background(), &model.StateHistory{
		ServerID:               serverID,
		Session:                state.Session,
		Track:                  state.Track,
		PlayerCount:            state.PlayerCount,
		DateCreated:            time.Now().UTC(),
		SessionStart:           state.SessionStart,
		SessionDurationMinutes: state.SessionDurationMinutes,
		SessionID:              sessionID,
	})
}

func (s *ServerService) updateSessionDuration(server *model.Server, sessionType model.TrackSession) {
	event, err := s.configService.GetEventConfig(server)
	if err != nil {
		event = &model.EventConfig{}
		logging.Error("Failed to get event config for server %d: %v", server.ID, err)
	}

	configuration, err := s.configService.GetConfiguration(server)
	if err != nil {
		configuration = &model.Configuration{}
		logging.Error("Failed to get configuration for server %d: %v", server.ID, err)
	}

	if instance, ok := s.instances.Load(server.ID); ok {
		serverInstance := instance.(*tracking.AccServerInstance)
		serverInstance.State.Track = event.Track
		serverInstance.State.MaxConnections = configuration.MaxConnections.ToInt()

		if serverInstance.State.Session != sessionType {
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

func (s *ServerService) GenerateServerPath(server *model.Server) {
	server.GenerateUUID()
	if server.ServiceName == "" {
		server.ServiceName = server.GenerateServiceName()
	}

	if env.IsDockerPlatform() {
		server.Path = filepath.Join(env.GetACCServersPath(), server.ServiceName)
		server.FromSteamCMD = true
	} else {
		steamCMDPath := env.GetSteamCMDDirPath()
		server.Path = server.GenerateServerPath(steamCMDPath)
		server.FromSteamCMD = true
	}
}

func (s *ServerService) handleStateChange(server *model.Server, state *model.ServerState) {
	s.updateSessionDuration(server, state.Session)

	s.apiService.statusCache.InvalidateStatus(server.ServiceName)

	if debouncer, exists := s.debouncers.Load(server.ID); exists {
		pending := debouncer.(*pendingState)
		pending.timer.Stop()
	}

	timer := time.NewTimer(5 * time.Minute)
	s.debouncers.Store(server.ID, &pendingState{
		timer: timer,
		state: state,
	})

	go func() {
		<-timer.C
		if debouncer, exists := s.debouncers.Load(server.ID); exists {
			pending := debouncer.(*pendingState)
			s.insertStateHistory(server.ID, pending.state)
			s.debouncers.Delete(server.ID)
		}
	}()

	if s.shouldInsertStateHistory(server.ID) {
		s.insertStateHistory(server.ID, state)
	}
}

func (s *ServerService) StartAccServerRuntime(server *model.Server) {
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

	serverIDStr := server.ID.String()
	s.configService.configCache.InvalidateServerCache(serverIDStr)

	s.updateSessionDuration(server, instance.State.Session)

	s.ensureLogTailing(server, instance)
}

func (s *ServerService) GetAll(ctx *fiber.Ctx, filter *model.ServerFilter) (*[]model.Server, error) {
	servers, err := s.repository.GetAll(ctx.UserContext(), filter)
	if err != nil {
		logging.Error("Failed to get servers: %v", err)
		return nil, err
	}

	// Collect server IDs for a single last-activity lookup.
	serverIDs := make([]uuid.UUID, len(*servers))
	for i, srv := range *servers {
		serverIDs[i] = srv.ID
	}
	lastActivities, laErr := s.activityLogRepo.GetLastActivityByServerIDs(ctx.UserContext(), serverIDs)
	if laErr != nil {
		logging.Error("Failed to get last activities for server list: %v", laErr)
	}

	for i := range *servers {
		server := &(*servers)[i]
		status, err := s.apiService.GetCachedStatus(server.ServiceName)
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
				server.State = serverInstance.State
			}
		}
		if lastActivities != nil {
			if info, ok := lastActivities[server.ID]; ok {
				server.LastActivity = info
			}
		}
	}

	return servers, nil
}

func (as *ServerService) GetById(ctx *fiber.Ctx, serverID uuid.UUID) (*model.Server, error) {
	server, err := as.repository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		return nil, err
	}
	status, err := as.apiService.GetCachedStatus(server.ServiceName)
	if err != nil {
		logging.Error("Failed to get cached status: %v", err)
	}
	server.Status = model.ParseServiceStatus(status)
	instance, ok := as.instances.Load(server.ID)
	if !ok {
		logging.Error("Unable to retrieve instance for server of ID: %s", server.ID)
	} else {
		serverInstance := instance.(*tracking.AccServerInstance)
		if serverInstance.State != nil {
			server.State = serverInstance.State
		}
	}

	return server, nil
}

func (s *ServerService) CreateServerAsync(ctx *fiber.Ctx, server *model.Server) error {
	logging.Info("create server start")
	if err := server.Validate(); err != nil {
		logging.Info("create server validation failed")
		return err
	}

	s.GenerateServerPath(server)

	bgCtx := context.Background()

	go func() {
		logging.Info("create server start background")
		if err := s.createServerBackground(bgCtx, server); err != nil {
			logging.Error("Async server creation failed for server %s: %v", server.ID, err)
			s.webSocketService.BroadcastError(server.ID, "Server creation failed", err.Error())
			s.webSocketService.BroadcastComplete(server.ID, false, fmt.Sprintf("Server creation failed: %v", err))
		}
	}()

	return nil
}

type createServerStep struct {
	stepType    model.ServerCreationStep
	important   bool
	callback    func() (string, error)
	description string
}

func (s *ServerService) createServerBackground(ctx context.Context, server *model.Server) error {
	var serverPort int
	var tcpPorts, udpPorts []int

	steps := []createServerStep{
		{
			stepType:    model.StepValidation,
			important:   true,
			description: "Server configuration validated successfully",
			callback: func() (string, error) {
				if err := server.Validate(); err != nil {
					return "", fmt.Errorf("validation failed: %v", err)
				}
				return "Server configuration validated successfully", nil
			},
		},
		{
			stepType:    model.StepDirectoryCreation,
			important:   true,
			description: "Server directories prepared",
			callback: func() (string, error) {
				return "Server directories prepared", nil
			},
		},
		{
			stepType:    model.StepSteamDownload,
			important:   true,
			description: "Server files downloaded successfully",
			callback: func() (string, error) {
				if err := s.steamService.InstallServerWithWebSocket(ctx, server.Path, &server.ID, s.webSocketService); err != nil {
					return "", fmt.Errorf("failed to install server: %v", err)
				}
				return "Server files downloaded successfully", nil
			},
		},
		{
			stepType:    model.StepConfigGeneration,
			important:   true,
			description: "",
			callback: func() (string, error) {
				var port int
				var err error

				if env.IsDockerPlatform() {
					port, err = s.portPool.Acquire(ctx)
					if err != nil {
						return "", fmt.Errorf("failed to acquire port from pool: %v", err)
					}
				} else {
					ports, err := network.FindAvailablePortRange(DefaultStartPort, RequiredPortCount)
					if err != nil {
						return "", fmt.Errorf("failed to find available ports: %v", err)
					}
					port = ports[0]
				}

				serverPort = port
				server.Port = port

				if err := s.updateServerPort(server, serverPort); err != nil {
					return "", fmt.Errorf("failed to update server configuration: %v", err)
				}

				return fmt.Sprintf("Server configuration generated (Port: %d)", serverPort), nil
			},
		},
		{
			stepType:    model.StepServiceCreation,
			important:   true,
			description: "",
			callback: func() (string, error) {
				server.Platform = env.GetPlatform()
				if err := s.runtime.Create(ctx, server); err != nil {
					return "", fmt.Errorf("failed to create server instance: %v", err)
				}
				return fmt.Sprintf("Server instance '%s' created successfully", server.ServiceName), nil
			},
		},
		{
			stepType:    model.StepFirewallRules,
			important:   false,
			description: "",
			callback: func() (string, error) {
				tcpPorts = []int{serverPort}
				udpPorts = []int{serverPort}
				if err := s.firewall.CreateServerRules(server.ServiceName, tcpPorts, udpPorts); err != nil {
					return "", fmt.Errorf("failed to create firewall rules: %v", err)
				}
				return fmt.Sprintf("Firewall rules configured for port %d", serverPort), nil
			},
		},
		{
			stepType:    model.StepDatabaseSave,
			important:   true,
			description: "Server saved to database successfully",
			callback: func() (string, error) {
				if err := s.repository.Insert(ctx, server); err != nil {
					return "", fmt.Errorf("failed to insert server into database: %v", err)
				}
				return "Server saved to database successfully", nil
			},
		},
	}

	for i, step := range steps {
		s.webSocketService.BroadcastStep(server.ID, step.stepType, model.StatusInProgress,
			model.GetStepDescription(step.stepType), "")

		successMessage, err := step.callback()
		if err != nil {
			s.webSocketService.BroadcastStep(server.ID, step.stepType, model.StatusFailed,
				"", err.Error())

			if step.important {
				s.rollbackSteps(ctx, server, steps[:i], tcpPorts, udpPorts)
				return err
			}
		}

		s.webSocketService.BroadcastStep(server.ID, step.stepType, model.StatusCompleted,
			successMessage, "")
	}

	s.StartAccServerRuntime(server)

	s.webSocketService.BroadcastStep(server.ID, model.StepCompleted, model.StatusCompleted,
		model.GetStepDescription(model.StepCompleted), "")

	s.webSocketService.BroadcastComplete(server.ID, true,
		fmt.Sprintf("Server '%s' created successfully on port %d", server.Name, serverPort))

	return nil
}

func (s *ServerService) rollbackSteps(ctx context.Context, server *model.Server, completedSteps []createServerStep, tcpPorts, udpPorts []int) {
	for i := len(completedSteps) - 1; i >= 0; i-- {
		step := completedSteps[i]
		switch step.stepType {
		case model.StepDatabaseSave:
			s.repository.Delete(ctx, server.ID)
		case model.StepFirewallRules:
			if len(tcpPorts) > 0 && len(udpPorts) > 0 {
				s.firewall.DeleteServerRules(server.ServiceName, tcpPorts, udpPorts)
			}
		case model.StepServiceCreation:
			s.runtime.Delete(ctx, server.ID)
		case model.StepSteamDownload:
			s.steamService.UninstallServer(server.Path)
		case model.StepDirectoryCreation:
			s.steamService.UninstallServer(server.Path)
		case model.StepConfigGeneration:
			if env.IsDockerPlatform() && server.Port > 0 {
				s.portPool.Release(ctx, server.Port)
			}
		}
	}
}

func (s *ServerService) DeleteServer(ctx *fiber.Ctx, serverID uuid.UUID) error {
	server, err := s.repository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		return fmt.Errorf("failed to get server details: %v", err)
	}

	if err := s.runtime.Delete(ctx.UserContext(), server.ID); err != nil {
		logging.Error("Failed to delete server instance: %v", err)
	}

	configuration, err := s.configService.GetConfiguration(server)
	if err != nil {
		logging.Error("Failed to get configuration for server %d: %v", server.ID, err)
	}
	if configuration != nil {
		tcpPorts := []int{configuration.TcpPort.ToInt()}
		udpPorts := []int{configuration.UdpPort.ToInt()}
		if err := s.firewall.DeleteServerRules(server.ServiceName, tcpPorts, udpPorts); err != nil {
			logging.Error("Failed to delete firewall rules: %v", err)
		}
	}

	if err := s.steamService.UninstallServer(server.Path); err != nil {
		logging.Error("Failed to uninstall server: %v", err)
	}

	if env.IsDockerPlatform() && server.Port > 0 {
		s.portPool.Release(ctx.UserContext(), server.Port)
	}

	if err := s.repository.Delete(ctx.UserContext(), serverID); err != nil {
		return fmt.Errorf("failed to delete server from database: %v", err)
	}

	s.logStreamer.Stop(server.ID)
	s.instances.Delete(server.ID)
	s.lastInsertTimes.Delete(server.ID)
	s.debouncers.Delete(server.ID)
	s.sessionIDs.Delete(server.ID)

	s.apiService.statusCache.InvalidateStatus(server.ServiceName)

	return nil
}

func (s *ServerService) configureFirewall(server *model.Server) error {
	ports, err := network.FindAvailablePortRange(DefaultStartPort, RequiredPortCount)
	if err != nil {
		return fmt.Errorf("failed to find available ports: %v", err)
	}

	serverPort := ports[0]
	tcpPorts := []int{serverPort}
	udpPorts := []int{serverPort}

	logging.Info("Configuring firewall for server %d with port %d", server.ID, serverPort)

	if err := s.firewall.UpdateServerRules(server.Name, tcpPorts, udpPorts); err != nil {
		return fmt.Errorf("failed to configure firewall: %v", err)
	}

	if err := s.updateServerPort(server, serverPort); err != nil {
		return fmt.Errorf("failed to update server configuration: %v", err)
	}

	return nil
}

func (s *ServerService) updateServerPort(server *model.Server, port int) error {
	config, err := s.configService.GetConfiguration(server)
	if err != nil {
		return fmt.Errorf("failed to load server configuration: %v", err)
	}

	config.TcpPort = model.IntString(port)
	config.UdpPort = model.IntString(port)
	config.RegisterToLobby = model.IntString(1)

	if err := s.configService.SaveConfiguration(server, config); err != nil {
		return fmt.Errorf("failed to save server configuration: %v", err)
	}

	if env.IsDockerPlatform() {
		if err := s.configService.SetIgnorePrematureDisconnects(server, 0); err != nil {
			logging.Error("Failed to set ignorePrematureDisconnects for Docker server: %v", err)
		}
	}

	return nil
}
