package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/env"
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/tracking"
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"acc-server-manager/local/utl/network"

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
	apiService       *ServiceControlService
	configService    *ConfigService
	steamService     *SteamService
	windowsService   *WindowsService
	firewallService  *FirewallService
	webSocketService *WebSocketService
	instances        sync.Map // Track instances per server
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
	if _, exists := s.logTailers.Load(server.ID); exists {
		return
	}

	go func() {
		logPath := filepath.Join(server.GetLogPath(), "server.log")
		tailer := tracking.NewLogTailer(logPath, instance.HandleLogLine)
		s.logTailers.Store(server.ID, tailer)

		tailer.Start()
	}()
}

func NewServerService(
	repository *repository.ServerRepository,
	stateHistoryRepo *repository.StateHistoryRepository,
	apiService *ServiceControlService,
	configService *ConfigService,
	steamService *SteamService,
	windowsService *WindowsService,
	firewallService *FirewallService,
	webSocketService *WebSocketService,
) *ServerService {
	service := &ServerService{
		repository:       repository,
		stateHistoryRepo: stateHistoryRepo,
		apiService:       apiService,
		configService:    configService,
		steamService:     steamService,
		windowsService:   windowsService,
		firewallService:  firewallService,
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
	steamCMDPath := env.GetSteamCMDDirPath()
	server.Path = server.GenerateServerPath(steamCMDPath)
	server.FromSteamCMD = true
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
	}

	return servers, nil
}

//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as *ServerService) GetById(ctx *fiber.Ctx, serverID uuid.UUID) (*model.Server, error) {
	server, err := as.repository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		return nil, err
	}
	status, err := as.apiService.GetCachedStatus(server.ServiceName)
	if err != nil {
		logging.Error(err.Error())
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
				ports, err := network.FindAvailablePortRange(DefaultStartPort, RequiredPortCount)
				if err != nil {
					return "", fmt.Errorf("failed to find available ports: %v", err)
				}

				serverPort = ports[0]

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
				execPath := filepath.Join(server.GetServerPath(), "accServer.exe")
				serverWorkingDir := filepath.Join(server.GetServerPath(), "server")
				if err := s.windowsService.CreateService(ctx, server.ServiceName, execPath, serverWorkingDir, nil); err != nil {
					return "", fmt.Errorf("failed to create Windows service: %v", err)
				}
				return fmt.Sprintf("Windows service '%s' created successfully", server.ServiceName), nil
			},
		},
		{
			stepType:    model.StepFirewallRules,
			important:   false,
			description: "",
			callback: func() (string, error) {
				s.configureFirewall(server)
				tcpPorts = []int{serverPort}
				udpPorts = []int{serverPort}
				if err := s.firewallService.CreateServerRules(server.ServiceName, tcpPorts, udpPorts); err != nil {
					return "", fmt.Errorf("failed to create firewall rules: %v", err)
				}
				return fmt.Sprintf("Firewall rules created for port %d", serverPort), nil
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
				s.firewallService.DeleteServerRules(server.ServiceName, tcpPorts, udpPorts)
			}
		case model.StepServiceCreation:
			s.windowsService.DeleteService(ctx, server.ServiceName)
		case model.StepSteamDownload:
			s.steamService.UninstallServer(server.Path)
		}
	}
}

func (s *ServerService) DeleteServer(ctx *fiber.Ctx, serverID uuid.UUID) error {
	server, err := s.repository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		return fmt.Errorf("failed to get server details: %v", err)
	}

	if err := s.windowsService.DeleteService(ctx.UserContext(), server.ServiceName); err != nil {
		logging.Error("Failed to delete Windows service: %v", err)
	}

	configuration, err := s.configService.GetConfiguration(server)
	if err != nil {
		logging.Error("Failed to get configuration for server %d: %v", server.ID, err)
	}
	tcpPorts := []int{configuration.TcpPort.ToInt()}
	udpPorts := []int{configuration.UdpPort.ToInt()}
	if err := s.firewallService.DeleteServerRules(server.ServiceName, tcpPorts, udpPorts); err != nil {
		logging.Error("Failed to delete firewall rules: %v", err)
	}

	if err := s.steamService.UninstallServer(server.Path); err != nil {
		logging.Error("Failed to uninstall server: %v", err)
	}

	if err := s.repository.Delete(ctx.UserContext(), serverID); err != nil {
		return fmt.Errorf("failed to delete server from database: %v", err)
	}

	if tailer, exists := s.logTailers.Load(server.ID); exists {
		tailer.(*tracking.LogTailer).Stop()
		s.logTailers.Delete(server.ID)
	}
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

	if err := s.firewallService.UpdateServerRules(server.Name, tcpPorts, udpPorts); err != nil {
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

	if err := s.configService.SaveConfiguration(server, config); err != nil {
		return fmt.Errorf("failed to save server configuration: %v", err)
	}

	return nil
}
