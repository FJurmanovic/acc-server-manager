package tests

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/configs"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TestHelper struct {
	DB       *gorm.DB
	TempDir  string
	TestData *TestData
}

type TestData struct {
	ServerID     uuid.UUID
	Server       *model.Server
	ConfigFiles  map[string]string
	SampleConfig *model.Configuration
}

func SetTestEnv() {
	os.Setenv("APP_SECRET", "test-secret-key-for-testing-123456")
	os.Setenv("APP_SECRET_CODE", "test-code-for-testing-123456789012")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-for-testing-123456789012345678901234567890")
	os.Setenv("ACCESS_KEY", "test-access-key-for-testing")
	os.Setenv("TESTING_ENV", "true")

	configs.Init()
}

func NewTestHelper(t *testing.T) *TestHelper {
	SetTestEnv()

	tempDir := t.TempDir()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(
		&model.Server{},
		&model.Config{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.StateHistory{},
	)

	if !db.Migrator().HasTable(&model.StateHistory{}) {
		err = db.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	testData := createTestData(t, tempDir)

	return &TestHelper{
		DB:       db,
		TempDir:  tempDir,
		TestData: testData,
	}
}

func createTestData(t *testing.T, tempDir string) *TestData {
	serverID := uuid.New()

	server := &model.Server{
		ID:           serverID,
		Name:         "Test Server",
		Path:         filepath.Join(tempDir, "server"),
		ServiceName:  "ACC-Server-Test",
		Status:       model.StatusStopped,
		DateCreated:  time.Now(),
		FromSteamCMD: false,
	}

	serverConfigDir := filepath.Join(tempDir, "server", "cfg")
	if err := os.MkdirAll(serverConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create server config directory: %v", err)
	}

	configFiles := map[string]string{
		"configuration.json": `{
			"udpPort": "9231",
			"tcpPort": "9232",
			"maxConnections": "30",
			"lanDiscovery": "1",
			"registerToLobby": "1",
			"configVersion": "1"
		}`,
		"settings.json": `{
			"serverName": "Test ACC Server",
			"adminPassword": "admin123",
			"carGroup": "GT3",
			"trackMedalsRequirement": "0",
			"safetyRatingRequirement": "30",
			"racecraftRatingRequirement": "30",
			"password": "",
			"spectatorPassword": "",
			"maxCarSlots": "30",
			"dumpLeaderboards": "1",
			"isRaceLocked": "0",
			"randomizeTrackWhenEmpty": "0",
			"centralEntryListPath": "",
			"allowAutoDQ": "1",
			"shortFormationLap": "0",
			"formationLapType": "3",
			"ignorePrematureDisconnects": "1"
		}`,
		"event.json": `{
			"track": "spa",
			"preRaceWaitingTimeSeconds": "80",
			"sessionOverTimeSeconds": "120",
			"ambientTemp": "26",
			"cloudLevel": 0.3,
			"rain": 0.0,
			"weatherRandomness": "1",
			"postQualySeconds": "10",
			"postRaceSeconds": "30",
			"simracerWeatherConditions": "0",
			"isFixedConditionQualification": "0",
			"sessions": [
				{
					"hourOfDay": "10",
					"dayOfWeekend": "1",
					"timeMultiplier": "1",
					"sessionType": "P",
					"sessionDurationMinutes": "10"
				},
				{
					"hourOfDay": "12",
					"dayOfWeekend": "1",
					"timeMultiplier": "1",
					"sessionType": "Q",
					"sessionDurationMinutes": "10"
				},
				{
					"hourOfDay": "14",
					"dayOfWeekend": "1",
					"timeMultiplier": "1",
					"sessionType": "R",
					"sessionDurationMinutes": "25"
				}
			]
		}`,
		"assistRules.json": `{
			"stabilityControlLevelMax": "0",
			"disableAutosteer": "1",
			"disableAutoLights": "0",
			"disableAutoWiper": "0",
			"disableAutoEngineStart": "0",
			"disableAutoPitLimiter": "0",
			"disableAutoGear": "0",
			"disableAutoClutch": "0",
			"disableIdealLine": "0"
		}`,
		"eventRules.json": `{
			"qualifyStandingType": "1",
			"pitWindowLengthSec": "600",
			"driverStIntStringTimeSec": "300",
			"mandatoryPitstopCount": "0",
			"maxTotalDrivingTime": "0",
			"isRefuellingAllowedInRace": 0,
			"isRefuellingTimeFixed": 0,
			"isMandatoryPitstopRefuellingRequired": 0,
			"isMandatoryPitstopTyreChangeRequired": 0,
			"isMandatoryPitstopSwapDriverRequired": 0,
			"tyreSetCount": "0"
		}`,
	}

	sampleConfig := &model.Configuration{
		UdpPort:         model.IntString(9231),
		TcpPort:         model.IntString(9232),
		MaxConnections:  model.IntString(30),
		LanDiscovery:    model.IntString(1),
		RegisterToLobby: model.IntString(1),
		ConfigVersion:   model.IntString(1),
	}

	return &TestData{
		ServerID:     serverID,
		Server:       server,
		ConfigFiles:  configFiles,
		SampleConfig: sampleConfig,
	}
}

func (th *TestHelper) CreateTestConfigFiles() error {
	serverConfigDir := filepath.Join(th.TestData.Server.Path, "cfg")

	for filename, content := range th.TestData.ConfigFiles {
		filePath := filepath.Join(serverConfigDir, filename)

		utf16Content, err := EncodeUTF16LEBOM([]byte(content))
		if err != nil {
			return err
		}

		if err := os.WriteFile(filePath, utf16Content, 0644); err != nil {
			return err
		}
	}

	return nil
}

func (th *TestHelper) CreateMalformedConfigFile(filename string) error {
	serverConfigDir := filepath.Join(th.TestData.Server.Path, "cfg")
	filePath := filepath.Join(serverConfigDir, filename)

	malformedJSON := `{
		"udpPort": "9231",
		"tcpPort": "9232"
		"maxConnections": "30"
	}`

	return os.WriteFile(filePath, []byte(malformedJSON), 0644)
}

func (th *TestHelper) RemoveConfigFile(filename string) error {
	serverConfigDir := filepath.Join(th.TestData.Server.Path, "cfg")
	filePath := filepath.Join(serverConfigDir, filename)
	return os.Remove(filePath)
}

func (th *TestHelper) InsertTestServer() error {
	return th.DB.Create(th.TestData.Server).Error
}

func (th *TestHelper) CreateContext() context.Context {
	return context.Background()
}

func (th *TestHelper) CreateFiberCtx() *fiber.Ctx {
	app := fiber.New()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	ctx.Locals("original-request", req)
	return ctx
}

func (th *TestHelper) ReleaseFiberCtx(app *fiber.App, ctx *fiber.Ctx) {
	if app != nil && ctx != nil {
		app.ReleaseCtx(ctx)
	}
}

func (th *TestHelper) Cleanup() {
	if sqlDB, err := th.DB.DB(); err == nil {
		sqlDB.Close()
	}

}

func LoadTestEnvFile() error {
	return godotenv.Load()
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func AssertError(t *testing.T, err error, expectedMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error containing '%s', got no error", expectedMsg)
	}
	if expectedMsg != "" && err.Error() != expectedMsg {
		t.Fatalf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil {
		t.Fatalf("Expected non-nil value, got nil")
	}
}

func AssertNil(t *testing.T, value interface{}) {
	t.Helper()
	if value != nil {
		switch v := value.(type) {
		case *interface{}:
			if v == nil || *v == nil {
				return
			}
		case interface{}:
			if v == nil {
				return
			}
		}
		t.Fatalf("Expected nil value, got %v", value)
	}
}

func EncodeUTF16LEBOM(input []byte) ([]byte, error) {
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
	return transformBytes(encoder.NewEncoder(), input)
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

func ErrorForTesting(message string) error {
	return errors.New(message)
}
