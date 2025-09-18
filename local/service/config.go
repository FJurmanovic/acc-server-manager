package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/logging"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/qjebbs/go-jsons"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

const (
	ConfigurationJson = "configuration.json"
	AssistRulesJson   = "assistRules.json"
	EventJson         = "event.json"
	EventRulesJson    = "eventRules.json"
	SettingsJson      = "settings.json"
)

var decodeMap = map[string]func(string) (interface{}, error){
	ConfigurationJson: func(f string) (interface{}, error) {
		return readAndDecode[model.Configuration](f, ConfigurationJson)
	},
	AssistRulesJson: func(f string) (interface{}, error) {
		return readAndDecode[model.AssistRules](f, AssistRulesJson)
	},
	EventJson: func(f string) (interface{}, error) {
		return readAndDecode[model.EventConfig](f, EventJson)
	},
	EventRulesJson: func(f string) (interface{}, error) {
		return readAndDecode[model.EventRules](f, EventRulesJson)
	},
	SettingsJson: func(f string) (interface{}, error) {
		return readAndDecode[model.ServerSettings](f, SettingsJson)
	},
}

func DecodeFileName(fileName string) func(path string) (interface{}, error) {
	if decoder, ok := decodeMap[fileName]; ok {
		return decoder
	}
	return nil
}

func mustDecode[T any](fileName, path string) (T, error) {
	result, err := DecodeFileName(fileName)(path)
	if err != nil {
		var zero T
		return zero, err
	}
	return result.(T), nil
}

type ConfigService struct {
	repository       *repository.ConfigRepository
	serverRepository *repository.ServerRepository
	serverService    *ServerService
	configCache      *model.ServerConfigCache
}

func NewConfigService(repository *repository.ConfigRepository, serverRepository *repository.ServerRepository) *ConfigService {
	logging.Debug("Initializing ConfigService with 5m expiration and 1s throttle")
	return &ConfigService{
		repository:       repository,
		serverRepository: serverRepository,
		configCache: model.NewServerConfigCache(model.CacheConfig{
			ExpirationTime: 5 * time.Minute,
			ThrottleTime:   1 * time.Second,
			DefaultStatus:  model.StatusUnknown,
		}),
	}
}

func (as *ConfigService) SetServerService(serverService *ServerService) {
	as.serverService = serverService
}

//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as *ConfigService) UpdateConfig(ctx *fiber.Ctx, body *map[string]interface{}) (*model.Config, error) {
	serverID := ctx.Locals("serverId").(string)
	configFile := ctx.Params("file")
	override := ctx.QueryBool("override", false)

	return as.updateConfigInternal(ctx.UserContext(), serverID, configFile, body, override)
}

func (as *ConfigService) updateConfigFiles(ctx context.Context, server *model.Server, configFile string, body *map[string]interface{}, override bool) ([]byte, []byte, error) {
	if server == nil {
		logging.Error("Server not found")
		return nil, nil, fmt.Errorf("server not found")
	}

	configPath := filepath.Join(server.GetConfigPath(), configFile)
	oldData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			dir := filepath.Dir(configPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, nil, err
			}
			if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
				return nil, nil, err
			}
			oldData = []byte("{}")
		} else {
			return nil, nil, err
		}
	}

	oldDataUTF8, err := DecodeUTF16LEBOM(oldData)
	if err != nil {
		return nil, nil, err
	}

	newData, err := json.Marshal(&body)
	if err != nil {
		return nil, nil, err
	}

	if !override {
		newData, err = jsons.Merge(oldDataUTF8, newData)
		if err != nil {
			return nil, nil, err
		}
	}
	newData, err = common.IndentJson(newData)
	if err != nil {
		return nil, nil, err
	}

	newDataUTF16, err := EncodeUTF16LEBOM(newData)
	if err != nil {
		return nil, nil, err
	}

	if err := os.WriteFile(configPath, newDataUTF16, 0644); err != nil {
		return nil, nil, err
	}

	return oldDataUTF8, newData, nil
}

func (as *ConfigService) updateConfigInternal(ctx context.Context, serverID string, configFile string, body *map[string]interface{}, override bool) (*model.Config, error) {
	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		logging.Error("Invalid server ID format: %v", err)
		return nil, fmt.Errorf("invalid server ID format")
	}

	server, err := as.serverRepository.GetByID(ctx, serverUUID)
	if err != nil {
		logging.Error("Server not found")
		return nil, fmt.Errorf("server not found")
	}

	oldDataUTF8, newData, err := as.updateConfigFiles(ctx, server, configFile, body, override)
	if err != nil {
		return nil, err
	}

	as.configCache.InvalidateServerCache(serverID)

	as.serverService.StartAccServerRuntime(server)
	return as.repository.UpdateConfig(ctx, &model.Config{
		ServerID:   serverUUID,
		ConfigFile: configFile,
		OldConfig:  string(oldDataUTF8),
		NewConfig:  string(newData),
		ChangedAt:  time.Now(),
	}), nil
}

//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as *ConfigService) GetConfig(ctx *fiber.Ctx) (interface{}, error) {
	serverIDStr := ctx.Params("id")
	configFile := ctx.Params("file")

	logging.Debug("Getting config for server ID: %s, file: %s", serverIDStr, configFile)

	server, err := as.serverRepository.GetByID(ctx.UserContext(), serverIDStr)

	if err != nil {
		logging.Error("Server not found")
		return nil, fiber.NewError(404, "Server not found")
	}
	return as.getConfigFile(server, configFile)
}

func (as *ConfigService) getConfigFile(server *model.Server, configFile string) (interface{}, error) {
	serverIDStr := server.ID.String()
	switch configFile {
	case ConfigurationJson:
		if cached, ok := as.configCache.GetConfiguration(serverIDStr); ok {
			logging.Debug("Returning cached configuration for server ID: %s", serverIDStr)
			return *cached, nil
		}
	case AssistRulesJson:
		if cached, ok := as.configCache.GetAssistRules(serverIDStr); ok {
			logging.Debug("Returning cached assist rules for server ID: %s", serverIDStr)
			return *cached, nil
		}
	case EventJson:
		if cached, ok := as.configCache.GetEvent(serverIDStr); ok {
			logging.Debug("Returning cached event config for server ID: %s", serverIDStr)
			return *cached, nil
		}
	case EventRulesJson:
		if cached, ok := as.configCache.GetEventRules(serverIDStr); ok {
			logging.Debug("Returning cached event rules for server ID: %s", serverIDStr)
			return *cached, nil
		}
	case SettingsJson:
		if cached, ok := as.configCache.GetSettings(serverIDStr); ok {
			logging.Debug("Returning cached settings for server ID: %s", serverIDStr)
			return *cached, nil
		}
	}

	logging.Debug("Cache miss for server ID: %s, file: %s - loading from disk", serverIDStr, configFile)

	configPath := server.GetConfigPath()
	decoder := DecodeFileName(configFile)
	if decoder == nil {
		return nil, errors.New("invalid config file")
	}

	config, err := decoder(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			logging.Debug("Config file not found, creating default for server ID: %s, file: %s", serverIDStr, configFile)
			switch configFile {
			case ConfigurationJson:
				return model.Configuration{}, nil
			case AssistRulesJson:
				return model.AssistRules{}, nil
			case EventJson:
				return model.EventConfig{}, nil
			case EventRulesJson:
				return model.EventRules{}, nil
			case SettingsJson:
				return model.ServerSettings{}, nil
			}
		}
		return nil, err
	}

	switch configFile {
	case ConfigurationJson:
		as.configCache.UpdateConfiguration(serverIDStr, config.(model.Configuration))
	case AssistRulesJson:
		as.configCache.UpdateAssistRules(serverIDStr, config.(model.AssistRules))
	case EventJson:
		as.configCache.UpdateEvent(serverIDStr, config.(model.EventConfig))
	case EventRulesJson:
		as.configCache.UpdateEventRules(serverIDStr, config.(model.EventRules))
	case SettingsJson:
		as.configCache.UpdateSettings(serverIDStr, config.(model.ServerSettings))
	}

	logging.Debug("Successfully loaded and cached config for server ID: %s, file: %s", serverIDStr, configFile)
	return config, nil
}

func (as *ConfigService) GetConfigs(ctx *fiber.Ctx) (*model.Configurations, error) {
	serverID := ctx.Params("id")

	server, err := as.serverRepository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		logging.Error("Server not found")
		return nil, fiber.NewError(404, "Server not found")
	}

	return as.LoadConfigs(server)
}

func (as *ConfigService) LoadConfigs(server *model.Server) (*model.Configurations, error) {
	serverIDStr := server.ID.String()
	logging.Info("Loading configs for server ID: %s at path: %s", serverIDStr, server.GetConfigPath())

	settingsConf, err := as.getConfigFile(server, SettingsJson)
	if err != nil {
		return nil, err
	}
	eventRulesConf, err := as.getConfigFile(server, EventRulesJson)
	if err != nil {
		return nil, err
	}
	eventConf, err := as.getConfigFile(server, EventJson)
	if err != nil {
		return nil, err
	}
	assistRulesConf, err := as.getConfigFile(server, AssistRulesJson)
	if err != nil {
		return nil, err
	}
	configurationConf, err := as.getConfigFile(server, ConfigurationJson)
	if err != nil {
		return nil, err
	}
	configs := &model.Configurations{
		Settings:      settingsConf.(model.ServerSettings),
		EventRules:    eventRulesConf.(model.EventRules),
		Event:         eventConf.(model.EventConfig),
		AssistRules:   assistRulesConf.(model.AssistRules),
		Configuration: configurationConf.(model.Configuration),
	}

	logging.Info("Successfully loaded all configs for server %s", serverIDStr)
	return configs, nil
}

func readAndDecode[T interface{}](path string, configFile string) (T, error) {
	settings, err := readFile(path, configFile)
	var zero T
	if err != nil {
		return zero, err
	}
	decodedsettings, err := DecodeToMap[T](settings)
	if err != nil {
		return zero, err
	}

	return decodedsettings, nil
}

func readFile(path string, configFile string) ([]byte, error) {
	configPath := filepath.Join(path, configFile)
	oldData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	return oldData, nil
}

func EncodeUTF16LEBOM(input []byte) ([]byte, error) {
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
	return transformBytes(encoder.NewEncoder(), input)
}

func DecodeUTF16LEBOM(input []byte) ([]byte, error) {
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
	return transformBytes(decoder.NewDecoder(), input)
}

func DecodeToMap[T interface{}](input []byte) (T, error) {
	var zero T
	if input == nil {
		return zero, fmt.Errorf("cannot decode nil input")
	}
	configUTF8 := new(T)
	decoded, err := DecodeUTF16LEBOM(input)
	if err != nil {
		return zero, fmt.Errorf("failed to decode UTF16: %v", err)
	}

	err = json.Unmarshal(decoded, configUTF8)
	if err != nil {
		return zero, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return *configUTF8, nil
}

func transformBytes(t transform.Transformer, input []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := transform.NewWriter(&buf, t)

	if _, err := io.Copy(w, bytes.NewReader(input)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (as *ConfigService) GetEventConfig(server *model.Server) (*model.EventConfig, error) {
	serverIDStr := server.ID.String()
	if cached, ok := as.configCache.GetEvent(serverIDStr); ok {
		return cached, nil
	}

	event, err := mustDecode[model.EventConfig](EventJson, server.GetConfigPath())
	if err != nil {
		return nil, err
	}
	as.configCache.UpdateEvent(serverIDStr, event)
	return &event, nil
}

func (as *ConfigService) GetConfiguration(server *model.Server) (*model.Configuration, error) {
	serverIDStr := server.ID.String()
	if cached, ok := as.configCache.GetConfiguration(serverIDStr); ok {
		return cached, nil
	}

	config, err := mustDecode[model.Configuration](ConfigurationJson, server.GetConfigPath())
	if err != nil {
		return nil, err
	}
	as.configCache.UpdateConfiguration(serverIDStr, config)
	return &config, nil
}

func (as *ConfigService) SaveConfiguration(server *model.Server, config *model.Configuration) error {
	configMap := make(map[string]interface{})
	configBytes, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %v", err)
	}
	if err := json.Unmarshal(configBytes, &configMap); err != nil {
		return fmt.Errorf("failed to unmarshal configuration: %v", err)
	}

	_, _, err = as.updateConfigFiles(context.Background(), server, ConfigurationJson, &configMap, true)
	return err
}
