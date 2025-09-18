package mocks

import (
	"acc-server-manager/local/model"
	"context"
	"errors"

	"github.com/google/uuid"
)

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

func (m *MockConfigRepository) SetShouldFailUpdate(shouldFail bool) {
	m.shouldFailUpdate = shouldFail
}

func (m *MockConfigRepository) GetConfig(serverID uuid.UUID, configFile string) *model.Config {
	key := serverID.String() + "_" + configFile
	return m.configs[key]
}

type MockServerRepository struct {
	servers       map[uuid.UUID]*model.Server
	shouldFailGet bool
}

func NewMockServerRepository() *MockServerRepository {
	return &MockServerRepository{
		servers: make(map[uuid.UUID]*model.Server),
	}
}

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

func (m *MockServerRepository) AddServer(server *model.Server) {
	m.servers[server.ID] = server
}

func (m *MockServerRepository) SetShouldFailGet(shouldFail bool) {
	m.shouldFailGet = shouldFail
}

type MockServerService struct {
	startRuntimeCalled bool
	startRuntimeServer *model.Server
}

func NewMockServerService() *MockServerService {
	return &MockServerService{}
}

func (m *MockServerService) StartAccServerRuntime(server *model.Server) {
	m.startRuntimeCalled = true
	m.startRuntimeServer = server
}

func (m *MockServerService) WasStartRuntimeCalled() bool {
	return m.startRuntimeCalled
}

func (m *MockServerService) GetStartRuntimeServer() *model.Server {
	return m.startRuntimeServer
}

func (m *MockServerService) Reset() {
	m.startRuntimeCalled = false
	m.startRuntimeServer = nil
}
