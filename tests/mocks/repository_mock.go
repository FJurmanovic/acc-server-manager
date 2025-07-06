package mocks

import (
	"acc-server-manager/local/model"
	"context"
	"errors"

	"github.com/google/uuid"
)

// MockConfigRepository provides a mock implementation of ConfigRepository
type MockConfigRepository struct {
	configs          map[string]*model.Config
	shouldFailGet    bool
	shouldFailUpdate bool
}

func NewMockConfigRepository() *MockConfigRepository {
	return &MockConfigRepository{
		configs: make(map[string]*model.Config),
	}
}

// UpdateConfig mocks the UpdateConfig method
func (m *MockConfigRepository) UpdateConfig(ctx context.Context, config *model.Config) *model.Config {
	if m.shouldFailUpdate {
		return nil
	}

	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}

	key := config.ServerID.String() + "_" + config.ConfigFile
	m.configs[key] = config
	return config
}

// SetShouldFailUpdate configures the mock to fail on UpdateConfig calls
func (m *MockConfigRepository) SetShouldFailUpdate(shouldFail bool) {
	m.shouldFailUpdate = shouldFail
}

// GetConfig retrieves a config by server ID and config file
func (m *MockConfigRepository) GetConfig(serverID uuid.UUID, configFile string) *model.Config {
	key := serverID.String() + "_" + configFile
	return m.configs[key]
}

// MockServerRepository provides a mock implementation of ServerRepository
type MockServerRepository struct {
	servers       map[uuid.UUID]*model.Server
	shouldFailGet bool
}

func NewMockServerRepository() *MockServerRepository {
	return &MockServerRepository{
		servers: make(map[uuid.UUID]*model.Server),
	}
}

// GetByID mocks the GetByID method
func (m *MockServerRepository) GetByID(ctx context.Context, id interface{}) (*model.Server, error) {
	if m.shouldFailGet {
		return nil, errors.New("server not found")
	}

	var serverID uuid.UUID
	var err error

	switch v := id.(type) {
	case string:
		serverID, err = uuid.Parse(v)
		if err != nil {
			return nil, errors.New("invalid server ID format")
		}
	case uuid.UUID:
		serverID = v
	default:
		return nil, errors.New("invalid server ID type")
	}

	server, exists := m.servers[serverID]
	if !exists {
		return nil, errors.New("server not found")
	}

	return server, nil
}

// AddServer adds a server to the mock repository
func (m *MockServerRepository) AddServer(server *model.Server) {
	m.servers[server.ID] = server
}

// SetShouldFailGet configures the mock to fail on GetByID calls
func (m *MockServerRepository) SetShouldFailGet(shouldFail bool) {
	m.shouldFailGet = shouldFail
}

// MockServerService provides a mock implementation of ServerService
type MockServerService struct {
	startRuntimeCalled bool
	startRuntimeServer *model.Server
}

func NewMockServerService() *MockServerService {
	return &MockServerService{}
}

// StartAccServerRuntime mocks the StartAccServerRuntime method
func (m *MockServerService) StartAccServerRuntime(server *model.Server) {
	m.startRuntimeCalled = true
	m.startRuntimeServer = server
}

// WasStartRuntimeCalled returns whether StartAccServerRuntime was called
func (m *MockServerService) WasStartRuntimeCalled() bool {
	return m.startRuntimeCalled
}

// GetStartRuntimeServer returns the server passed to StartAccServerRuntime
func (m *MockServerService) GetStartRuntimeServer() *model.Server {
	return m.startRuntimeServer
}

// Reset resets the mock state
func (m *MockServerService) Reset() {
	m.startRuntimeCalled = false
	m.startRuntimeServer = nil
}
