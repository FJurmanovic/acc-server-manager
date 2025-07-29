package tests

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/configs"
	"acc-server-manager/local/utl/jwt"
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

// TestHelper provides utilities for testing
type TestHelper struct {
	DB       *gorm.DB
	TempDir  string
	TestData *TestData
}

// TestData contains common test data structures
type TestData struct {
	ServerID     uuid.UUID
	Server       *model.Server
	ConfigFiles  map[string]string
	SampleConfig *model.Configuration
}

// SetTestEnv sets the required environment variables for tests
func SetTestEnv() {
	// Set required environment variables for testing
	os.Setenv("APP_SECRET", "test-secret-key-for-testing-123456")
	os.Setenv("APP_SECRET_CODE", "test-code-for-testing-123456789012")
	os.Setenv("ENCRYPTION_KEY", "12345678901234567890123456789012")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-for-testing-123456789012345678901234567890")
	os.Setenv("ACCESS_KEY", "test-access-key-for-testing")
	// Set test-specific environment variables
	os.Setenv("TESTING_ENV", "true") // Used to bypass

	configs.Init()
	jwt.Init()
}

// NewTestHelper creates a new test helper with in-memory database
func NewTestHelper(t *testing.T) *TestHelper {
	// Set required environment variables
	SetTestEnv()

	// Create temporary directory for test files
	tempDir := t.TempDir()

	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Suppress SQL logs in tests
	})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&model.Server{},
		&model.Config{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.StateHistory{},
	)

	// Explicitly ensure tables exist with correct structure
	if !db.Migrator().HasTable(&model.StateHistory{}) {
		err = db.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Create test data
	testData := createTestData(t, tempDir)

	return &TestHelper{
		DB:       db,
		TempDir:  tempDir,
		TestData: testData,
	}
}

// createTestData creates common test data structures
func createTestData(t *testing.T, tempDir string) *TestData {
	serverID := uuid.New()

	// Create sample server
	server := &model.Server{
		ID:           serverID,
		Name:         "Test Server",
		Path:         filepath.Join(tempDir, "server"),
		ServiceName:  "ACC-Server-Test",
		Status:       model.StatusStopped,
		DateCreated:  time.Now(),
		FromSteamCMD: false,
	}

	// Create server directory
	serverConfigDir := filepath.Join(tempDir, "server", "cfg")
	if err := os.MkdirAll(serverConfigDir, 0755); err != nil {
		t.Fatalf("Failed to create server config directory: %v", err)
	}

	// Sample configuration files content
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

	// Sample configuration struct
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

// CreateTestConfigFiles creates actual config files in the test directory
func (th *TestHelper) CreateTestConfigFiles() error {
	serverConfigDir := filepath.Join(th.TestData.Server.Path, "cfg")

	for filename, content := range th.TestData.ConfigFiles {
		filePath := filepath.Join(serverConfigDir, filename)

		// Encode content to UTF-16 LE BOM format as expected by the application
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

// CreateMalformedConfigFile creates a config file with invalid JSON
func (th *TestHelper) CreateMalformedConfigFile(filename string) error {
	serverConfigDir := filepath.Join(th.TestData.Server.Path, "cfg")
	filePath := filepath.Join(serverConfigDir, filename)

	malformedJSON := `{
		"udpPort": "9231",
		"tcpPort": "9232"
		"maxConnections": "30"  // Missing comma - invalid JSON
	}`

	return os.WriteFile(filePath, []byte(malformedJSON), 0644)
}

// RemoveConfigFile removes a config file to simulate missing file scenarios
func (th *TestHelper) RemoveConfigFile(filename string) error {
	serverConfigDir := filepath.Join(th.TestData.Server.Path, "cfg")
	filePath := filepath.Join(serverConfigDir, filename)
	return os.Remove(filePath)
}

// InsertTestServer inserts the test server into the database
func (th *TestHelper) InsertTestServer() error {
	return th.DB.Create(th.TestData.Server).Error
}

// CreateContext creates a test context
func (th *TestHelper) CreateContext() context.Context {
	return context.Background()
}

// CreateFiberCtx creates a fiber.Ctx for testing
func (th *TestHelper) CreateFiberCtx() *fiber.Ctx {
	// Create app and request for fiber context
	app := fiber.New()
	// Create a dummy request that doesn't depend on external http objects
	req := httptest.NewRequest("GET", "/", nil)
	// Create the fiber context from real request/response
	ctx := app.AcquireCtx(&fasthttp.RequestCtx{})
	// Store the original request for release later
	ctx.Locals("original-request", req)
	// Return the context which can be safely used in tests
	return ctx
}

// ReleaseFiberCtx properly releases a fiber context created with CreateFiberCtx
func (th *TestHelper) ReleaseFiberCtx(app *fiber.App, ctx *fiber.Ctx) {
	if app != nil && ctx != nil {
		app.ReleaseCtx(ctx)
	}
}

// Cleanup performs cleanup operations after tests
func (th *TestHelper) Cleanup() {
	// Close database connection
	if sqlDB, err := th.DB.DB(); err == nil {
		sqlDB.Close()
	}

	// Temporary directory is automatically cleaned up by t.TempDir()
}

// LoadTestEnvFile loads environment variables from a .env file for testing
func LoadTestEnvFile() error {
	// Try to load from .env file
	return godotenv.Load()
}

// AssertNoError is a helper function to check for errors in tests
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

// AssertError is a helper function to check for expected errors
func AssertError(t *testing.T, err error, expectedMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error containing '%s', got no error", expectedMsg)
	}
	if expectedMsg != "" && err.Error() != expectedMsg {
		t.Fatalf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if expected != actual {
		t.Fatalf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()
	if value == nil {
		t.Fatalf("Expected non-nil value, got nil")
	}
}

// AssertNil checks if a value is nil
func AssertNil(t *testing.T, value interface{}) {
	t.Helper()
	if value != nil {
		// Special handling for interface values that contain nil but aren't nil themselves
		// For example, (*jwt.Claims)(nil) is not equal to nil, but it contains nil
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

// EncodeUTF16LEBOM encodes UTF-8 bytes to UTF-16 LE BOM format
func EncodeUTF16LEBOM(input []byte) ([]byte, error) {
	encoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
	return transformBytes(encoder.NewEncoder(), input)
}

// transformBytes applies a transform to input bytes
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

// ErrorForTesting creates an error for testing purposes
func ErrorForTesting(message string) error {
	return errors.New(message)
}
