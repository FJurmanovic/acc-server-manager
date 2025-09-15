package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/service"
	"acc-server-manager/tests"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigService_GetConfiguration_ValidFile(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create test config files
	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	// Create repositories and service
	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	// Test GetConfiguration
	config, err := configService.GetConfiguration(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, config)

	// Verify the result is the expected configuration
	tests.AssertEqual(t, model.IntString(9231), config.UdpPort)
	tests.AssertEqual(t, model.IntString(9232), config.TcpPort)
	tests.AssertEqual(t, model.IntString(30), config.MaxConnections)
	tests.AssertEqual(t, model.IntString(1), config.LanDiscovery)
	tests.AssertEqual(t, model.IntString(1), config.RegisterToLobby)
	tests.AssertEqual(t, model.IntString(1), config.ConfigVersion)
}

func TestConfigService_GetConfiguration_MissingFile(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create server directory but no config files
	serverConfigDir := filepath.Join(helper.TestData.Server.Path, "cfg")
	err := os.MkdirAll(serverConfigDir, 0755)
	tests.AssertNoError(t, err)

	// Create repositories and service
	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	// Test GetConfiguration for missing file
	config, err := configService.GetConfiguration(helper.TestData.Server)
	if err == nil {
		t.Fatal("Expected error for missing file, got nil")
	}
	if config != nil {
		t.Fatal("Expected nil config, got non-nil")
	}
}

func TestConfigService_GetEventConfig_ValidFile(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create test config files
	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	// Create repositories and service
	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	// Test GetEventConfig
	eventConfig, err := configService.GetEventConfig(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, eventConfig)

	// Verify the result is the expected event configuration
	tests.AssertEqual(t, "spa", eventConfig.Track)
	tests.AssertEqual(t, model.IntString(80), eventConfig.PreRaceWaitingTimeSeconds)
	tests.AssertEqual(t, model.IntString(120), eventConfig.SessionOverTimeSeconds)
	tests.AssertEqual(t, model.IntString(26), eventConfig.AmbientTemp)
	tests.AssertEqual(t, float64(0.3), eventConfig.CloudLevel)
	tests.AssertEqual(t, float64(0.0), eventConfig.Rain)

	// Verify sessions
	tests.AssertEqual(t, 3, len(eventConfig.Sessions))
	if len(eventConfig.Sessions) > 0 {
		tests.AssertEqual(t, model.SessionPractice, eventConfig.Sessions[0].SessionType)
		tests.AssertEqual(t, model.IntString(10), eventConfig.Sessions[0].SessionDurationMinutes)
	}
}

func TestConfigService_SaveConfiguration_Success(t *testing.T) {
	t.Skip("Temporarily disabled due to path issues")
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create test config files
	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	// Create repositories and service
	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	// Prepare new configuration
	newConfig := &model.Configuration{
		UdpPort:         model.IntString(9999),
		TcpPort:         model.IntString(10000),
		MaxConnections:  model.IntString(40),
		LanDiscovery:    model.IntString(0),
		RegisterToLobby: model.IntString(1),
		ConfigVersion:   model.IntString(2),
	}

	// Test SaveConfiguration
	err = configService.SaveConfiguration(helper.TestData.Server, newConfig)
	tests.AssertNoError(t, err)

	// Verify the configuration was saved
	configPath := filepath.Join(helper.TestData.Server.Path, "cfg", "configuration.json")
	fileContent, err := os.ReadFile(configPath)
	tests.AssertNoError(t, err)

	// Convert from UTF-16 to UTF-8 for verification
	utf8Content, err := service.DecodeUTF16LEBOM(fileContent)
	tests.AssertNoError(t, err)

	var savedConfig map[string]interface{}
	err = json.Unmarshal(utf8Content, &savedConfig)
	tests.AssertNoError(t, err)

	// Verify the saved values
	tests.AssertEqual(t, "9999", savedConfig["udpPort"])
	tests.AssertEqual(t, "10000", savedConfig["tcpPort"])
	tests.AssertEqual(t, "40", savedConfig["maxConnections"])
	tests.AssertEqual(t, "0", savedConfig["lanDiscovery"])
	tests.AssertEqual(t, "1", savedConfig["registerToLobby"])
	tests.AssertEqual(t, "2", savedConfig["configVersion"])
}

func TestConfigService_LoadConfigs_Success(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create test config files
	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	// Create repositories and service
	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	// Test LoadConfigs
	configs, err := configService.LoadConfigs(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, configs)

	// Verify all configurations are loaded
	tests.AssertEqual(t, model.IntString(9231), configs.Configuration.UdpPort)
	tests.AssertEqual(t, model.IntString(9232), configs.Configuration.TcpPort)
	tests.AssertEqual(t, "Test ACC Server", configs.Settings.ServerName)
	tests.AssertEqual(t, "admin123", configs.Settings.AdminPassword)
	tests.AssertEqual(t, "spa", configs.Event.Track)
	tests.AssertEqual(t, model.IntString(80), configs.Event.PreRaceWaitingTimeSeconds)
	tests.AssertEqual(t, model.IntString(0), configs.AssistRules.StabilityControlLevelMax)
	tests.AssertEqual(t, model.IntString(1), configs.AssistRules.DisableAutosteer)
	tests.AssertEqual(t, model.IntString(1), configs.EventRules.QualifyStandingType)
	tests.AssertEqual(t, model.IntString(600), configs.EventRules.PitWindowLengthSec)
}

func TestConfigService_LoadConfigs_MissingFiles(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create server directory but no config files
	serverConfigDir := filepath.Join(helper.TestData.Server.Path, "cfg")
	err := os.MkdirAll(serverConfigDir, 0755)
	tests.AssertNoError(t, err)

	// Create repositories and service
	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	// Test LoadConfigs with missing files
	configs, err := configService.LoadConfigs(helper.TestData.Server)
	if err == nil {
		t.Fatal("Expected error for missing files, got nil")
	}
	if configs != nil {
		t.Fatal("Expected nil configs, got non-nil")
	}
}

func TestConfigService_MalformedJSON(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create malformed config file
	err := helper.CreateMalformedConfigFile("configuration.json")
	tests.AssertNoError(t, err)

	// Create repositories and service
	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	// Test GetConfiguration with malformed JSON
	config, err := configService.GetConfiguration(helper.TestData.Server)
	if err == nil {
		t.Fatal("Expected error for malformed JSON, got nil")
	}
	if config != nil {
		t.Fatal("Expected nil config, got non-nil")
	}
}

func TestConfigService_UTF16_Encoding(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create test config files
	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	// Test UTF-16 encoding and decoding
	originalData := `{"udpPort": "9231", "tcpPort": "9232"}`

	// Encode to UTF-16 LE BOM
	encoded, err := service.EncodeUTF16LEBOM([]byte(originalData))
	tests.AssertNoError(t, err)

	// Decode back to UTF-8
	decoded, err := service.DecodeUTF16LEBOM(encoded)
	tests.AssertNoError(t, err)

	// Verify it matches original
	tests.AssertEqual(t, originalData, string(decoded))
}

func TestConfigService_DecodeFileName(t *testing.T) {
	// Test that all supported file names have decoders
	testCases := []string{
		"configuration.json",
		"assistRules.json",
		"event.json",
		"eventRules.json",
		"settings.json",
	}

	for _, filename := range testCases {
		t.Run(filename, func(t *testing.T) {
			decoder := service.DecodeFileName(filename)
			tests.AssertNotNil(t, decoder)
		})
	}

	// Test invalid filename
	decoder := service.DecodeFileName("invalid.json")
	if decoder != nil {
		t.Fatal("Expected nil decoder for invalid filename, got non-nil")
	}
}

func TestConfigService_IntString_Conversion(t *testing.T) {
	// Test IntString unmarshaling from string
	var intStr model.IntString

	// Test string input
	err := json.Unmarshal([]byte(`"123"`), &intStr)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 123, intStr.ToInt())
	tests.AssertEqual(t, "123", intStr.ToString())

	// Test int input
	err = json.Unmarshal([]byte(`456`), &intStr)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 456, intStr.ToInt())
	tests.AssertEqual(t, "456", intStr.ToString())

	// Test empty string
	err = json.Unmarshal([]byte(`""`), &intStr)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 0, intStr.ToInt())
	tests.AssertEqual(t, "0", intStr.ToString())
}

func TestConfigService_IntBool_Conversion(t *testing.T) {
	// Test IntBool unmarshaling from int
	var intBool model.IntBool

	// Test int input (1 = true)
	err := json.Unmarshal([]byte(`1`), &intBool)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 1, intBool.ToInt())
	tests.AssertEqual(t, true, intBool.ToBool())

	// Test int input (0 = false)
	err = json.Unmarshal([]byte(`0`), &intBool)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 0, intBool.ToInt())
	tests.AssertEqual(t, false, intBool.ToBool())

	// Test bool input (true)
	err = json.Unmarshal([]byte(`true`), &intBool)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 1, intBool.ToInt())
	tests.AssertEqual(t, true, intBool.ToBool())

	// Test bool input (false)
	err = json.Unmarshal([]byte(`false`), &intBool)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 0, intBool.ToInt())
	tests.AssertEqual(t, false, intBool.ToBool())
}

func TestConfigService_Caching_Configuration(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create test config files (already UTF-16 encoded)
	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	// Create repositories and service
	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	// First call - should load from disk
	config1, err := configService.GetConfiguration(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, config1)

	// Modify the file on disk with UTF-16 encoding
	configPath := filepath.Join(helper.TestData.Server.Path, "cfg", "configuration.json")
	modifiedContent := `{"udpPort": "5555", "tcpPort": "5556"}`
	utf16Modified, err := service.EncodeUTF16LEBOM([]byte(modifiedContent))
	tests.AssertNoError(t, err)

	err = os.WriteFile(configPath, utf16Modified, 0644)
	tests.AssertNoError(t, err)

	// Second call - should return cached result (not the modified file)
	config2, err := configService.GetConfiguration(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, config2)

	// Should still have the original cached values
	tests.AssertEqual(t, model.IntString(9231), config2.UdpPort)
	tests.AssertEqual(t, model.IntString(9232), config2.TcpPort)
}

func TestConfigService_Caching_EventConfig(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create test config files (already UTF-16 encoded)
	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	// Create repositories and service
	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	// First call - should load from disk
	event1, err := configService.GetEventConfig(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, event1)

	// Modify the file on disk with UTF-16 encoding
	configPath := filepath.Join(helper.TestData.Server.Path, "cfg", "event.json")
	modifiedContent := `{"track": "monza", "preRaceWaitingTimeSeconds": "60"}`
	utf16Modified, err := service.EncodeUTF16LEBOM([]byte(modifiedContent))
	tests.AssertNoError(t, err)

	err = os.WriteFile(configPath, utf16Modified, 0644)
	tests.AssertNoError(t, err)

	// Second call - should return cached result (not the modified file)
	event2, err := configService.GetEventConfig(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, event2)

	// Should still have the original cached values
	tests.AssertEqual(t, "spa", event2.Track)
	tests.AssertEqual(t, model.IntString(80), event2.PreRaceWaitingTimeSeconds)
}
