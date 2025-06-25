package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/logging"
	"acc-server-manager/local/utl/tracking"
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"acc-server-manager/local/utl/network"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultStartPort = 9600
	RequiredPortCount = 1 // Update this if ACC needs more ports
)

type ServerService struct {
	repository       *repository.ServerRepository
	stateHistoryRepo *repository.StateHistoryRepository
	apiService       *ApiService
	configService    *ConfigService
	steamService     *SteamService
	windowsService   *WindowsService
	firewallService  *FirewallService
	systemConfigService *SystemConfigService
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
	// Check if we already have a tailer
	if _, exists := s.logTailers.Load(server.ID); exists {
		return
	}

	// Start tailing in a goroutine that handles file creation/deletion
	go func() {
		logPath := filepath.Join(server.GetLogPath(), "server.log")
		tailer := tracking.NewLogTailer(logPath, instance.HandleLogLine)
		s.logTailers.Store(server.ID, tailer)
		
		// Start tailing and automatically handle file changes
		tailer.Start()
	}()
}

func NewServerService(
	repository *repository.ServerRepository,
	stateHistoryRepo *repository.StateHistoryRepository,
	apiService *ApiService,
	configService *ConfigService,
	steamService *SteamService,
	windowsService *WindowsService,
	firewallService *FirewallService,
	systemConfigService *SystemConfigService,
) *ServerService {
	service := &ServerService{
		repository:       repository,
		stateHistoryRepo: stateHistoryRepo,
		apiService:       apiService,
		configService:    configService,
		steamService:     steamService,
		windowsService:   windowsService,
		firewallService:  firewallService,
		systemConfigService: systemConfigService,
	}

	// Initialize server instances
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
	lastID, err := s.stateHistoryRepo.GetLastSessionID(context.Background(), serverID)
	if err != nil {
		logging.Error("Failed to get last session ID for server %d: %v", serverID, err)
		return 1 // Return 1 as fallback
	}
	return lastID + 1
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

func (s *ServerService) GenerateServerPath(server *model.Server) {
	// Get the base steamcmd path
	steamCMDPath, err := s.systemConfigService.GetSteamCMDDirPath(context.Background())
	if err != nil {
		logging.Error("Failed to get steamcmd path: %v", err)
		return
	}

	server.Path = server.GenerateServerPath(steamCMDPath)
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
				server.State = serverInstance.State
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
			server.State = serverInstance.State
		}
	}

	return server, nil
}

func (s *ServerService) CreateServer(ctx *fiber.Ctx, server *model.Server) error {
	// Validate basic server configuration
	if err := server.Validate(); err != nil {
		return err
	}

	// Install server using SteamCMD
	if err := s.steamService.InstallServer(ctx.UserContext(), server.GetServerPath()); err != nil {
		return fmt.Errorf("failed to install server: %v", err)
	}

	// Create Windows service with correct paths
	execPath := filepath.Join(server.GetServerPath(), "accServer.exe")
	serverWorkingDir := filepath.Join(server.GetServerPath(), "server")
	if err := s.windowsService.CreateService(ctx.UserContext(), server.ServiceName, execPath, serverWorkingDir, nil); err != nil {
		// Cleanup on failure
		s.steamService.UninstallServer(server.Path)
		return fmt.Errorf("failed to create Windows service: %v", err)
	}

	s.configureFirewall(server)
	ports, err := network.FindAvailablePortRange(DefaultStartPort, RequiredPortCount)
	if err != nil {
		return fmt.Errorf("failed to find available ports: %v", err)
	}

	// Use the first port for both TCP and UDP
	serverPort := ports[0]
	tcpPorts := []int{serverPort}
	udpPorts := []int{serverPort}
	if err := s.firewallService.CreateServerRules(server.ServiceName, tcpPorts, udpPorts); err != nil {
		// Cleanup on failure
		s.windowsService.DeleteService(ctx.UserContext(), server.ServiceName)
		s.steamService.UninstallServer(server.Path)
		return fmt.Errorf("failed to create firewall rules: %v", err)
	}

	// Update server configuration with the allocated port
	if err := s.updateServerPort(server, serverPort); err != nil {
		return fmt.Errorf("failed to update server configuration: %v", err)
	}

	// Insert server into database
	if err := s.repository.Insert(ctx.UserContext(), server); err != nil {
		// Cleanup on failure
		s.firewallService.DeleteServerRules(server.ServiceName, tcpPorts, udpPorts)
		s.windowsService.DeleteService(ctx.UserContext(), server.ServiceName)
		s.steamService.UninstallServer(server.Path)
		return fmt.Errorf("failed to insert server into database: %v", err)
	}

	// Initialize server runtime
	s.StartAccServerRuntime(server)

	return nil
}

func (s *ServerService) DeleteServer(ctx *fiber.Ctx, serverID int) error {
	// Get server details
	server, err := s.repository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		return fmt.Errorf("failed to get server details: %v", err)
	}

	// Stop and remove Windows service
	if err := s.windowsService.DeleteService(ctx.UserContext(), server.ServiceName); err != nil {
		logging.Error("Failed to delete Windows service: %v", err)
	}


	// Remove firewall rules
	configuration, err := s.configService.GetConfiguration(server)
	if err != nil {
		logging.Error("Failed to get configuration for server %d: %v", server.ID, err)
	}
	tcpPorts := []int{configuration.TcpPort.ToInt()}
	udpPorts := []int{configuration.UdpPort.ToInt()}
	if err := s.firewallService.DeleteServerRules(server.ServiceName, tcpPorts, udpPorts); err != nil {
		logging.Error("Failed to delete firewall rules: %v", err)
	}

	// Uninstall server files
	if err := s.steamService.UninstallServer(server.Path); err != nil {
		logging.Error("Failed to uninstall server: %v", err)
	}

	// Remove from database
	if err := s.repository.Delete(ctx.UserContext(), serverID); err != nil {
		return fmt.Errorf("failed to delete server from database: %v", err)
	}

	// Cleanup runtime resources
	if tailer, exists := s.logTailers.Load(server.ID); exists {
		tailer.(*tracking.LogTailer).Stop()
		s.logTailers.Delete(server.ID)
	}
	s.instances.Delete(server.ID)
	s.lastInsertTimes.Delete(server.ID)
	s.debouncers.Delete(server.ID)
	s.sessionIDs.Delete(server.ID)

	return nil
}

func (s *ServerService) UpdateServer(ctx *fiber.Ctx, server *model.Server) error {
	// Validate server configuration
	if err := server.Validate(); err != nil {
		return err
	}

	// Get existing server details
	existingServer, err := s.repository.GetByID(ctx.UserContext(), int(server.ID))
	if err != nil {
		return fmt.Errorf("failed to get existing server details: %v", err)
	}

	// Update server files if path changed
	if existingServer.Path != server.Path {
		if err := s.steamService.InstallServer(ctx.UserContext(), server.Path); err != nil {
			return fmt.Errorf("failed to install server to new location: %v", err)
		}
		// Clean up old installation
		if err := s.steamService.UninstallServer(existingServer.Path); err != nil {
			logging.Error("Failed to remove old server installation: %v", err)
		}
	}

	// Update Windows service if necessary
	if existingServer.ServiceName != server.ServiceName || existingServer.Path != server.Path {
		execPath := filepath.Join(server.GetServerPath(), "accServer.exe")
		serverWorkingDir := server.GetServerPath()
		if err := s.windowsService.UpdateService(ctx.UserContext(), server.ServiceName, execPath, serverWorkingDir, nil); err != nil {
			return fmt.Errorf("failed to update Windows service: %v", err)
		}
	}

	// Update firewall rules if service name changed
	if existingServer.ServiceName != server.ServiceName {
		if err := s.configureFirewall(server); err != nil {
			return fmt.Errorf("failed to update firewall rules: %v", err)
		}
	}

	// Update database record
	if err := s.repository.Update(ctx.UserContext(), server); err != nil {
		return fmt.Errorf("failed to update server in database: %v", err)
	}

	// Restart server runtime
	s.StartAccServerRuntime(server)

	return nil
}

func (s *ServerService) configureFirewall(server *model.Server) error {
	// Find available ports for the server
	ports, err := network.FindAvailablePortRange(DefaultStartPort, RequiredPortCount)
	if err != nil {
		return fmt.Errorf("failed to find available ports: %v", err)
	}

	// Use the first port for both TCP and UDP
	serverPort := ports[0]
	tcpPorts := []int{serverPort}
	udpPorts := []int{serverPort}

	logging.Info("Configuring firewall for server %d with port %d", server.ID, serverPort)

	// Configure firewall rules
	if err := s.firewallService.UpdateServerRules(server.Name, tcpPorts, udpPorts); err != nil {
		return fmt.Errorf("failed to configure firewall: %v", err)
	}

	// Update server configuration with the allocated port
	if err := s.updateServerPort(server, serverPort); err != nil {
		return fmt.Errorf("failed to update server configuration: %v", err)
	}

	return nil
}

func (s *ServerService) updateServerPort(server *model.Server, port int) error {
	// Load current configuration
	config, err := s.configService.GetConfiguration(server)
	if err != nil {
		return fmt.Errorf("failed to load server configuration: %v", err)
	}

	config.TcpPort = model.IntString(port)
	config.UdpPort = model.IntString(port)

	// Save the updated configuration
	if err := s.configService.SaveConfiguration(server, config); err != nil {
		return fmt.Errorf("failed to save server configuration: %v", err)
	}

	return nil
}