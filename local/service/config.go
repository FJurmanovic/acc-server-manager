package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/qjebbs/go-jsons"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type ConfigService struct {
	repository       *repository.ConfigRepository
	serverRepository *repository.ServerRepository
}

func NewConfigService(repository *repository.ConfigRepository, serverRepository *repository.ServerRepository) *ConfigService {
	return &ConfigService{
		repository:       repository,
		serverRepository: serverRepository,
	}
}

// UpdateConfig
// Updates physical config file and caches it in database.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as ConfigService) UpdateConfig(ctx *fiber.Ctx, body *map[string]interface{}) (*model.Config, error) {
	serverID, _ := ctx.ParamsInt("id")
	configFile := ctx.Params("file")
	merge := ctx.QueryBool("merge")

	server := as.serverRepository.GetFirst(ctx.UserContext(), serverID)

	if server == nil {
		return nil, fiber.NewError(404, "Server not found")
	}

	// Read existing config
	configPath := filepath.Join(server.ConfigPath, "\\server\\cfg", configFile)
	oldData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	oldDataUTF8, err := DecodeUTF16LEBOM(oldData)
	if err != nil {
		return nil, err
	}

	// Write new config
	newData, err := json.MarshalIndent(&body, "", "  ")
	if err != nil {
		return nil, err
	}

	if merge {
		newData, err = jsons.Merge(oldDataUTF8, newData)
		if err != nil {
			return nil, err
		}
	}

	newDataUTF16, err := EncodeUTF16LEBOM(newData)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(configPath, newDataUTF16, 0644); err != nil {
		return nil, err
	}

	// Log change
	return as.repository.UpdateConfig(ctx.UserContext(), &model.Config{
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
func (as ConfigService) GetConfig(ctx *fiber.Ctx) (map[string]interface{}, error) {
	serverID, _ := ctx.ParamsInt("id")
	configFile := ctx.Params("file")

	server := as.serverRepository.GetFirst(ctx.UserContext(), serverID)

	if server == nil {
		return nil, fiber.NewError(404, "Server not found")
	}

	config, err := readFile(server.ConfigPath, configFile)

	if err != nil {
		return nil, err
	}

	decoded, err := DecodeToMap(config)
	if err != nil {
		return nil, err
	}

	return decoded, nil
}

// GetConfigs
// Gets physical config file and caches it in database.
//
//	   	Args:
//	   		context.Context: Application context
//		Returns:
//			string: Application version
func (as ConfigService) GetConfigs(ctx *fiber.Ctx) (*model.Configurations, error) {
	serverID, _ := ctx.ParamsInt("id")

	server := as.serverRepository.GetFirst(ctx.UserContext(), serverID)

	if server == nil {
		return nil, fiber.NewError(404, "Server not found")
	}

	configuration, err := readFile(server.ConfigPath, "configuration.json")
	if err != nil {
		return nil, err
	}
	decodedconfiguration, err := DecodeToMap(configuration)
	if err != nil {
		return nil, err
	}

	entrylist, err := readFile(server.ConfigPath, "entrylist.json")
	if err != nil {
		return nil, err
	}
	decodedentrylist, err := DecodeToMap(entrylist)
	if err != nil {
		return nil, err
	}

	event, err := readFile(server.ConfigPath, "event.json")
	if err != nil {
		return nil, err
	}
	decodedevent, err := DecodeToMap(event)
	if err != nil {
		return nil, err
	}

	eventRules, err := readFile(server.ConfigPath, "eventRules.json")
	if err != nil {
		return nil, err
	}
	decodedeventRules, err := DecodeToMap(eventRules)
	if err != nil {
		return nil, err
	}

	settings, err := readFile(server.ConfigPath, "settings.json")
	if err != nil {
		return nil, err
	}
	decodedsettings, err := DecodeToMap(settings)
	if err != nil {
		return nil, err
	}

	return &model.Configurations{
		Configuration: decodedconfiguration,
		Event:         decodedevent,
		EventRules:    decodedeventRules,
		Settings:      decodedsettings,
		Entrylist:     decodedentrylist,
	}, nil
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

func DecodeToMap(input []byte) (map[string]interface{}, error) {
	if input == nil {
		return nil, nil
	}
	configUTF8 := new(map[string]interface{})
	decoded, err := DecodeUTF16LEBOM(input)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decoded, configUTF8)
	if err != nil {
		return nil, err
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
