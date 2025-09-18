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
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	config, err := configService.GetConfiguration(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, config)

	tests.AssertEqual(t, model.IntString(9231), config.UdpPort)
	tests.AssertEqual(t, model.IntString(9232), config.TcpPort)
	tests.AssertEqual(t, model.IntString(30), config.MaxConnections)
	tests.AssertEqual(t, model.IntString(1), config.LanDiscovery)
	tests.AssertEqual(t, model.IntString(1), config.RegisterToLobby)
	tests.AssertEqual(t, model.IntString(1), config.ConfigVersion)
}

func TestConfigService_GetEventConfig_ValidFile(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	eventConfig, err := configService.GetEventConfig(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, eventConfig)

	tests.AssertEqual(t, "spa", eventConfig.Track)
	tests.AssertEqual(t, model.IntString(80), eventConfig.PreRaceWaitingTimeSeconds)
	tests.AssertEqual(t, model.IntString(120), eventConfig.SessionOverTimeSeconds)
	tests.AssertEqual(t, model.IntString(26), eventConfig.AmbientTemp)
	tests.AssertEqual(t, float64(0.3), eventConfig.CloudLevel)
	tests.AssertEqual(t, float64(0.0), eventConfig.Rain)

	tests.AssertEqual(t, 3, len(eventConfig.Sessions))
	if len(eventConfig.Sessions) > 0 {
		tests.AssertEqual(t, model.SessionPractice, eventConfig.Sessions[0].SessionType)
		tests.AssertEqual(t, model.IntString(10), eventConfig.Sessions[0].SessionDurationMinutes)
	}
}

func TestConfigService_SaveConfiguration_Success(t *testing.T) {
	t.Skip("Temporarily disabled due to path issues")
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	newConfig := &model.Configuration{
		UdpPort:         model.IntString(9999),
		TcpPort:         model.IntString(10000),
		MaxConnections:  model.IntString(40),
		LanDiscovery:    model.IntString(0),
		RegisterToLobby: model.IntString(1),
		ConfigVersion:   model.IntString(2),
	}

	err = configService.SaveConfiguration(helper.TestData.Server, newConfig)
	tests.AssertNoError(t, err)

	configPath := filepath.Join(helper.TestData.Server.Path, "cfg", "configuration.json")
	fileContent, err := os.ReadFile(configPath)
	tests.AssertNoError(t, err)

	utf8Content, err := service.DecodeUTF16LEBOM(fileContent)
	tests.AssertNoError(t, err)

	var savedConfig map[string]interface{}
	err = json.Unmarshal(utf8Content, &savedConfig)
	tests.AssertNoError(t, err)

	tests.AssertEqual(t, "9999", savedConfig["udpPort"])
	tests.AssertEqual(t, "10000", savedConfig["tcpPort"])
	tests.AssertEqual(t, "40", savedConfig["maxConnections"])
	tests.AssertEqual(t, "0", savedConfig["lanDiscovery"])
	tests.AssertEqual(t, "1", savedConfig["registerToLobby"])
	tests.AssertEqual(t, "2", savedConfig["configVersion"])
}

func TestConfigService_LoadConfigs_Success(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	configs, err := configService.LoadConfigs(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, configs)

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

func TestConfigService_MalformedJSON(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	err := helper.CreateMalformedConfigFile("configuration.json")
	tests.AssertNoError(t, err)

	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	config, err := configService.GetConfiguration(helper.TestData.Server)
	if err == nil {
		t.Fatal("Expected error for malformed JSON, got nil")
	}
	if config != nil {
		t.Fatal("Expected nil config, got non-nil")
	}
}

func TestConfigService_UTF16_Encoding(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	originalData := `{"udpPort": "9231", "tcpPort": "9232"}`

	encoded, err := service.EncodeUTF16LEBOM([]byte(originalData))
	tests.AssertNoError(t, err)

	decoded, err := service.DecodeUTF16LEBOM(encoded)
	tests.AssertNoError(t, err)

	tests.AssertEqual(t, originalData, string(decoded))
}

func TestConfigService_DecodeFileName(t *testing.T) {
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

	decoder := service.DecodeFileName("invalid.json")
	if decoder != nil {
		t.Fatal("Expected nil decoder for invalid filename, got non-nil")
	}
}

func TestConfigService_IntString_Conversion(t *testing.T) {
	var intStr model.IntString

	err := json.Unmarshal([]byte(`"123"`), &intStr)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 123, intStr.ToInt())
	tests.AssertEqual(t, "123", intStr.ToString())

	err = json.Unmarshal([]byte(`456`), &intStr)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 456, intStr.ToInt())
	tests.AssertEqual(t, "456", intStr.ToString())

	err = json.Unmarshal([]byte(`""`), &intStr)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 0, intStr.ToInt())
	tests.AssertEqual(t, "0", intStr.ToString())
}

func TestConfigService_IntBool_Conversion(t *testing.T) {
	var intBool model.IntBool

	err := json.Unmarshal([]byte(`1`), &intBool)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 1, intBool.ToInt())
	tests.AssertEqual(t, true, intBool.ToBool())

	err = json.Unmarshal([]byte(`0`), &intBool)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 0, intBool.ToInt())
	tests.AssertEqual(t, false, intBool.ToBool())

	err = json.Unmarshal([]byte(`true`), &intBool)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 1, intBool.ToInt())
	tests.AssertEqual(t, true, intBool.ToBool())

	err = json.Unmarshal([]byte(`false`), &intBool)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 0, intBool.ToInt())
	tests.AssertEqual(t, false, intBool.ToBool())
}

func TestConfigService_Caching_Configuration(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	config1, err := configService.GetConfiguration(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, config1)

	configPath := filepath.Join(helper.TestData.Server.Path, "cfg", "configuration.json")
	modifiedContent := `{"udpPort": "5555", "tcpPort": "5556"}`
	utf16Modified, err := service.EncodeUTF16LEBOM([]byte(modifiedContent))
	tests.AssertNoError(t, err)

	err = os.WriteFile(configPath, utf16Modified, 0644)
	tests.AssertNoError(t, err)

	config2, err := configService.GetConfiguration(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, config2)

	tests.AssertEqual(t, model.IntString(9231), config2.UdpPort)
	tests.AssertEqual(t, model.IntString(9232), config2.TcpPort)
}

func TestConfigService_Caching_EventConfig(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	err := helper.CreateTestConfigFiles()
	tests.AssertNoError(t, err)

	configRepo := repository.NewConfigRepository(helper.DB)
	serverRepo := repository.NewServerRepository(helper.DB)
	configService := service.NewConfigService(configRepo, serverRepo)

	event1, err := configService.GetEventConfig(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, event1)

	configPath := filepath.Join(helper.TestData.Server.Path, "cfg", "event.json")
	modifiedContent := `{"track": "monza", "preRaceWaitingTimeSeconds": "60"}`
	utf16Modified, err := service.EncodeUTF16LEBOM([]byte(modifiedContent))
	tests.AssertNoError(t, err)

	err = os.WriteFile(configPath, utf16Modified, 0644)
	tests.AssertNoError(t, err)

	event2, err := configService.GetEventConfig(helper.TestData.Server)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, event2)

	tests.AssertEqual(t, "spa", event2.Track)
	tests.AssertEqual(t, model.IntString(80), event2.PreRaceWaitingTimeSeconds)
}
