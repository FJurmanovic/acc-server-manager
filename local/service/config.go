package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/common"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
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
	return &ConfigService{
		repository:       repository,
		serverRepository: serverRepository,
		configCache:     model.NewServerConfigCache(model.CacheConfig{
			ExpirationTime: 5 * time.Minute,  // Cache configs for 5 minutes
			ThrottleTime:   1 * time.Second,  // Prevent rapid re-reads
			DefaultStatus:  model.StatusUnknown,
		}),
	}
}

func (as *ConfigService) SetServerService(serverService *ServerService) {
	as.serverService = serverService
}

// UpdateConfig
// Updates physical config file and caches it in database.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as ConfigService) UpdateConfig(ctx *fiber.Ctx, body *map[string]interface{}) (*model.Config, error) {
	serverID := ctx.Locals("serverId").(int)
	configFile := ctx.Params("file")
	override := ctx.QueryBool("override", false)

	server, err := as.serverRepository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		log.Print("Server not found")
		return nil, fiber.NewError(404, "Server not found")
	}

	// Read existing config
	configPath := filepath.Join(server.ConfigPath, "\\server\\cfg", configFile)
	oldData, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory if it doesn't exist
			dir := filepath.Dir(configPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, err
			}
			// Create empty JSON file
			if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
				return nil, err
			}
			oldData = []byte("{}")
		} else {
			return nil, err
		}
	}

	oldDataUTF8, err := DecodeUTF16LEBOM(oldData)
	if err != nil {
		return nil, err
	}

	// Write new config
	newData, err := json.Marshal(&body)
	if err != nil {
		return nil, err
	}

	if !override {
		newData, err = jsons.Merge(oldDataUTF8, newData)
		if err != nil {
			return nil, err
		}
	}
	newData, err = common.IndentJson(newData)
	if err != nil {
		return nil, err
	}

	newDataUTF16, err := EncodeUTF16LEBOM(newData)
	if err != nil {
		return nil, err
	}

	context := ctx.UserContext()

	if err := os.WriteFile(configPath, newDataUTF16, 0644); err != nil {
		return nil, err
	}

	// Invalidate all configs for this server since configs can be interdependent
	as.configCache.InvalidateServerCache(strconv.Itoa(serverID))

	as.serverService.StartAccServerRuntime(server)

	// Log change
	return as.repository.UpdateConfig(context, &model.Config{
		ServerID:   uint(serverID),
		ConfigFile: configFile,
		OldConfig:  string(oldDataUTF8),
		NewConfig:  string(newData),
		ChangedAt:  time.Now(),
	}), nil
}

// GetConfig
// Gets physical config file and caches it in database.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as ConfigService) GetConfig(ctx *fiber.Ctx) (interface{}, error) {
	serverID, _ := ctx.ParamsInt("id")
	configFile := ctx.Params("file")
	serverIDStr := strconv.Itoa(serverID)

	server, err := as.serverRepository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		log.Print("Server not found")
		return nil, fiber.NewError(404, "Server not found")
	}

	// Try to get from cache based on config file type
	switch configFile {
	case ConfigurationJson:
		if cached, ok := as.configCache.GetConfiguration(serverIDStr); ok {
			return cached, nil
		}
	case AssistRulesJson:
		if cached, ok := as.configCache.GetAssistRules(serverIDStr); ok {
			return cached, nil
		}
	case EventJson:
		if cached, ok := as.configCache.GetEvent(serverIDStr); ok {
			return cached, nil
		}
	case EventRulesJson:
		if cached, ok := as.configCache.GetEventRules(serverIDStr); ok {
			return cached, nil
		}
	case SettingsJson:
		if cached, ok := as.configCache.GetSettings(serverIDStr); ok {
			return cached, nil
		}
	}

	decoded, err := DecodeFileName(configFile)(server.ConfigPath)
	if err != nil {
		return nil, err
	}

	// Cache the result based on config file type
	switch configFile {
	case ConfigurationJson:
		if config, ok := decoded.(model.Configuration); ok {
			as.configCache.UpdateConfiguration(serverIDStr, config)
		}
	case AssistRulesJson:
		if rules, ok := decoded.(model.AssistRules); ok {
			as.configCache.UpdateAssistRules(serverIDStr, rules)
		}
	case EventJson:
		if event, ok := decoded.(model.EventConfig); ok {
			as.configCache.UpdateEvent(serverIDStr, event)
		}
	case EventRulesJson:
		if rules, ok := decoded.(model.EventRules); ok {
			as.configCache.UpdateEventRules(serverIDStr, rules)
		}
	case SettingsJson:
		if settings, ok := decoded.(model.ServerSettings); ok {
			as.configCache.UpdateSettings(serverIDStr, settings)
		}
	}

	return decoded, nil
}

// GetConfigs
// Gets all configurations for a server, using cache when possible.
func (as ConfigService) GetConfigs(ctx *fiber.Ctx) (*model.Configurations, error) {
	serverID, _ := ctx.ParamsInt("id")
	serverIDStr := strconv.Itoa(serverID)

	server, err := as.serverRepository.GetByID(ctx.UserContext(), serverID)
	if err != nil {
		log.Print("Server not found")
		return nil, fiber.NewError(404, "Server not found")
	}

	configs := &model.Configurations{}

	// Load configuration
	if cached, ok := as.configCache.GetConfiguration(serverIDStr); ok {
		configs.Configuration = *cached
	} else {
		config, err := mustDecode[model.Configuration](ConfigurationJson, server.ConfigPath)
		if err != nil {
			return nil, err
		}
		configs.Configuration = config
		as.configCache.UpdateConfiguration(serverIDStr, config)
	}

	// Load assist rules
	if cached, ok := as.configCache.GetAssistRules(serverIDStr); ok {
		configs.AssistRules = *cached
	} else {
		rules, err := mustDecode[model.AssistRules](AssistRulesJson, server.ConfigPath)
		if err != nil {
			return nil, err
		}
		configs.AssistRules = rules
		as.configCache.UpdateAssistRules(serverIDStr, rules)
	}

	// Load event config
	if cached, ok := as.configCache.GetEvent(serverIDStr); ok {
		configs.Event = *cached
	} else {
		event, err := mustDecode[model.EventConfig](EventJson, server.ConfigPath)
		if err != nil {
			return nil, err
		}
		configs.Event = event
		as.configCache.UpdateEvent(serverIDStr, event)
	}

	// Load event rules
	if cached, ok := as.configCache.GetEventRules(serverIDStr); ok {
		configs.EventRules = *cached
	} else {
		rules, err := mustDecode[model.EventRules](EventRulesJson, server.ConfigPath)
		if err != nil {
			return nil, err
		}
		configs.EventRules = rules
		as.configCache.UpdateEventRules(serverIDStr, rules)
	}

	// Load settings
	if cached, ok := as.configCache.GetSettings(serverIDStr); ok {
		configs.Settings = *cached
	} else {
		settings, err := mustDecode[model.ServerSettings](SettingsJson, server.ConfigPath)
		if err != nil {
			return nil, err
		}
		configs.Settings = settings
		as.configCache.UpdateSettings(serverIDStr, settings)
	}

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
	configPath := filepath.Join(path, "\\server\\cfg", configFile)
	oldData, err := os.ReadFile(configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	} else if errors.Is(err, os.ErrNotExist) {
		return nil, nil
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
		return zero, nil
	}
	configUTF8 := new(T)
	decoded, err := DecodeUTF16LEBOM(input)
	if err != nil {
		return zero, err
	}

	err = json.Unmarshal(decoded, configUTF8)
	if err != nil {
		return zero, err
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
