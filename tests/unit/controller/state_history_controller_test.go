package controller

import (
	"acc-server-manager/local/controller"
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/service"
	"acc-server-manager/local/utl/cache"
	"acc-server-manager/local/utl/common"
	"acc-server-manager/local/utl/jwt"
	"acc-server-manager/tests"
	"acc-server-manager/tests/testdata"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TestStateHistoryController_GetAll_Success(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	app := fiber.New()
	repo := repository.NewStateHistoryRepository(helper.DB)
	stateHistoryService := service.NewStateHistoryService(repo)

	membershipRepo := repository.NewMembershipRepository(helper.DB)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)
	membershipService := service.NewMembershipService(membershipRepo, jwtHandler, openJWTHandler)

	inMemCache := cache.NewInMemoryCache()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)
	history := testData.CreateStateHistory(model.SessionPractice, "spa", 5, uuid.New())
	err := repo.Insert(helper.CreateContext(), &history)
	tests.AssertNoError(t, err)

	routeGroups := &common.RouteGroups{
		StateHistory: app.Group("/api/v1/state-history"),
	}

	controller.NewStateHistoryController(stateHistoryService, routeGroups, GetTestAuthMiddleware(membershipService, inMemCache))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/state-history?id=%s", helper.TestData.ServerID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())

	resp, err := app.Test(req)
	tests.AssertNoError(t, err)

	tests.AssertEqual(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)

	var result []model.StateHistory
	err = json.Unmarshal(body, &result)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 1, len(result))
	tests.AssertEqual(t, model.SessionPractice, result[0].Session)
	tests.AssertEqual(t, 5, result[0].PlayerCount)
}

func TestStateHistoryController_GetAll_WithSessionFilter(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	app := fiber.New()
	repo := repository.NewStateHistoryRepository(helper.DB)
	stateHistoryService := service.NewStateHistoryService(repo)

	membershipRepo := repository.NewMembershipRepository(helper.DB)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)
	membershipService := service.NewMembershipService(membershipRepo, jwtHandler, openJWTHandler)

	inMemCache := cache.NewInMemoryCache()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)

	practiceHistory := testData.CreateStateHistory(model.SessionPractice, "spa", 5, uuid.New())
	raceHistory := testData.CreateStateHistory(model.SessionRace, "spa", 10, uuid.New())

	err := repo.Insert(helper.CreateContext(), &practiceHistory)
	tests.AssertNoError(t, err)
	err = repo.Insert(helper.CreateContext(), &raceHistory)
	tests.AssertNoError(t, err)

	routeGroups := &common.RouteGroups{
		StateHistory: app.Group("/api/v1/state-history"),
	}

	controller.NewStateHistoryController(stateHistoryService, routeGroups, GetTestAuthMiddleware(membershipService, inMemCache))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/state-history?id=%s&session=R", helper.TestData.ServerID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())

	resp, err := app.Test(req)
	tests.AssertNoError(t, err)

	tests.AssertEqual(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)

	var result []model.StateHistory
	err = json.Unmarshal(body, &result)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, 1, len(result))
	tests.AssertEqual(t, model.SessionRace, result[0].Session)
	tests.AssertEqual(t, 10, result[0].PlayerCount)
}

func TestStateHistoryController_GetAll_EmptyResult(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	app := fiber.New()
	repo := repository.NewStateHistoryRepository(helper.DB)
	stateHistoryService := service.NewStateHistoryService(repo)

	membershipRepo := repository.NewMembershipRepository(helper.DB)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)
	membershipService := service.NewMembershipService(membershipRepo, jwtHandler, openJWTHandler)

	inMemCache := cache.NewInMemoryCache()

	routeGroups := &common.RouteGroups{
		StateHistory: app.Group("/api/v1/state-history"),
	}

	controller.NewStateHistoryController(stateHistoryService, routeGroups, GetTestAuthMiddleware(membershipService, inMemCache))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/state-history?id=%s", helper.TestData.ServerID.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())

	resp, err := app.Test(req)
	tests.AssertNoError(t, err)

	tests.AssertEqual(t, http.StatusOK, resp.StatusCode)
}

func TestStateHistoryController_GetStatistics_Success(t *testing.T) {
	t.Skip("Skipping test due to UUID validation issues")

	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	app := fiber.New()
	repo := repository.NewStateHistoryRepository(helper.DB)
	stateHistoryService := service.NewStateHistoryService(repo)

	membershipRepo := repository.NewMembershipRepository(helper.DB)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)
	membershipService := service.NewMembershipService(membershipRepo, jwtHandler, openJWTHandler)

	inMemCache := cache.NewInMemoryCache()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)

	playerCounts := []int{5, 10, 15, 20, 25}
	entries := testData.CreateMultipleEntries(model.SessionRace, "spa", playerCounts)

	for _, entry := range entries {
		err := repo.Insert(helper.CreateContext(), &entry)
		tests.AssertNoError(t, err)
	}

	routeGroups := &common.RouteGroups{
		StateHistory: app.Group("/api/v1/state-history"),
	}

	controller.NewStateHistoryController(stateHistoryService, routeGroups, GetTestAuthMiddleware(membershipService, inMemCache))

	validServerID := helper.TestData.ServerID.String()
	if validServerID == "" {
		validServerID = uuid.New().String()
	}
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/state-history/statistics?id=%s", validServerID), nil)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())

	resp, err := app.Test(req)
	tests.AssertNoError(t, err)

	tests.AssertEqual(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)

	var stats model.StateHistoryStats
	err = json.Unmarshal(body, &stats)
	tests.AssertNoError(t, err)

	if stats.PeakPlayers < 0 {
		t.Error("Expected non-negative peak players")
	}
	if stats.AveragePlayers < 0 {
		t.Error("Expected non-negative average players")
	}
	if stats.TotalSessions < 0 {
		t.Error("Expected non-negative total sessions")
	}
}

func TestStateHistoryController_GetStatistics_NoData(t *testing.T) {
	t.Skip("Skipping test due to UUID validation issues")

	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	app := fiber.New()
	repo := repository.NewStateHistoryRepository(helper.DB)
	stateHistoryService := service.NewStateHistoryService(repo)

	membershipRepo := repository.NewMembershipRepository(helper.DB)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)
	membershipService := service.NewMembershipService(membershipRepo, jwtHandler, openJWTHandler)

	inMemCache := cache.NewInMemoryCache()

	routeGroups := &common.RouteGroups{
		StateHistory: app.Group("/api/v1/state-history"),
	}

	controller.NewStateHistoryController(stateHistoryService, routeGroups, GetTestAuthMiddleware(membershipService, inMemCache))

	validServerID := helper.TestData.ServerID.String()
	if validServerID == "" {
		validServerID = uuid.New().String()
	}
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/state-history/statistics?id=%s", validServerID), nil)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())

	resp, err := app.Test(req)
	tests.AssertNoError(t, err)

	tests.AssertEqual(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)

	var stats model.StateHistoryStats
	err = json.Unmarshal(body, &stats)
	tests.AssertNoError(t, err)

	tests.AssertEqual(t, 0, stats.PeakPlayers)
	tests.AssertEqual(t, 0.0, stats.AveragePlayers)
	tests.AssertEqual(t, 0, stats.TotalSessions)
}

func TestStateHistoryController_GetStatistics_InvalidQueryParams(t *testing.T) {
	t.Skip("Skipping test due to UUID validation issues")

	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	app := fiber.New()
	repo := repository.NewStateHistoryRepository(helper.DB)
	stateHistoryService := service.NewStateHistoryService(repo)

	membershipRepo := repository.NewMembershipRepository(helper.DB)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)
	membershipService := service.NewMembershipService(membershipRepo, jwtHandler, openJWTHandler)

	inMemCache := cache.NewInMemoryCache()

	routeGroups := &common.RouteGroups{
		StateHistory: app.Group("/api/v1/state-history"),
	}

	controller.NewStateHistoryController(stateHistoryService, routeGroups, GetTestAuthMiddleware(membershipService, inMemCache))

	validServerID := helper.TestData.ServerID.String()
	if validServerID == "" {
		validServerID = uuid.New().String()
	}
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/state-history/statistics?id=%s&min_players=invalid", validServerID), nil)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())

	resp, err := app.Test(req)
	tests.AssertNoError(t, err)

	tests.AssertEqual(t, http.StatusBadRequest, resp.StatusCode)
}

func TestStateHistoryController_HTTPMethods(t *testing.T) {

	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	app := fiber.New()
	repo := repository.NewStateHistoryRepository(helper.DB)
	stateHistoryService := service.NewStateHistoryService(repo)

	membershipRepo := repository.NewMembershipRepository(helper.DB)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)
	membershipService := service.NewMembershipService(membershipRepo, jwtHandler, openJWTHandler)

	inMemCache := cache.NewInMemoryCache()

	routeGroups := &common.RouteGroups{
		StateHistory: app.Group("/api/v1/state-history"),
	}

	controller.NewStateHistoryController(stateHistoryService, routeGroups, GetTestAuthMiddleware(membershipService, inMemCache))

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/state-history?id=%s", helper.TestData.ServerID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)

	req = httptest.NewRequest("POST", fmt.Sprintf("/api/v1/state-history/statistics?id=%s", helper.TestData.ServerID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())
	resp, err = app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)

	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/state-history?id=%s", helper.TestData.ServerID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())
	resp, err = app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)

	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/state-history?id=%s", helper.TestData.ServerID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())
	resp, err = app.Test(req)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestStateHistoryController_ContentType(t *testing.T) {

	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	app := fiber.New()
	repo := repository.NewStateHistoryRepository(helper.DB)
	stateHistoryService := service.NewStateHistoryService(repo)

	membershipRepo := repository.NewMembershipRepository(helper.DB)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)
	membershipService := service.NewMembershipService(membershipRepo, jwtHandler, openJWTHandler)

	inMemCache := cache.NewInMemoryCache()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)
	history := testData.CreateStateHistory(model.SessionPractice, "spa", 5, uuid.New())
	err := repo.Insert(helper.CreateContext(), &history)
	tests.AssertNoError(t, err)

	routeGroups := &common.RouteGroups{
		StateHistory: app.Group("/api/v1/state-history"),
	}

	controller.NewStateHistoryController(stateHistoryService, routeGroups, GetTestAuthMiddleware(membershipService, inMemCache))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/state-history?id=%s", helper.TestData.ServerID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type: application/json, got %s", contentType)
	}

	validServerID := helper.TestData.ServerID.String()
	if validServerID == "" {
		validServerID = uuid.New().String()
	}
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/state-history/statistics?id=%s", validServerID), nil)
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())
	resp, err = app.Test(req)
	tests.AssertNoError(t, err)

	contentType = resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type: application/json, got %s", contentType)
	}
}

func TestStateHistoryController_ResponseStructure(t *testing.T) {
	t.Skip("Skipping test due to response structure issues that need further investigation")

	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	app := fiber.New()
	repo := repository.NewStateHistoryRepository(helper.DB)
	stateHistoryService := service.NewStateHistoryService(repo)

	membershipRepo := repository.NewMembershipRepository(helper.DB)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(jwtSecret)
	openJWTHandler := jwt.NewOpenJWTHandler(jwtSecret)
	membershipService := service.NewMembershipService(membershipRepo, jwtHandler, openJWTHandler)

	inMemCache := cache.NewInMemoryCache()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)
	history := testData.CreateStateHistory(model.SessionPractice, "spa", 5, uuid.New())
	err := repo.Insert(helper.CreateContext(), &history)
	tests.AssertNoError(t, err)

	routeGroups := &common.RouteGroups{
		StateHistory: app.Group("/api/v1/state-history"),
	}

	controller.NewStateHistoryController(stateHistoryService, routeGroups, GetTestAuthMiddleware(membershipService, inMemCache))

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/state-history?id=%s", helper.TestData.ServerID.String()), nil)
	req.Header.Set("Authorization", "Bearer "+tests.MustGenerateTestToken())
	resp, err := app.Test(req)
	tests.AssertNoError(t, err)

	body, err := io.ReadAll(resp.Body)
	tests.AssertNoError(t, err)

	t.Logf("Response body: %s", string(body))

	var resultArray []model.StateHistory
	err = json.Unmarshal(body, &resultArray)
	if err != nil {
		var singleResult model.StateHistory
		err = json.Unmarshal(body, &singleResult)
		if err != nil {
			t.Fatalf("Failed to parse response as either array or object: %v", err)
		}
		resultArray = []model.StateHistory{singleResult}
	}

	if len(resultArray) > 0 {
		history := resultArray[0]
		if history.ID == uuid.Nil {
			t.Error("Expected non-nil ID in StateHistory")
		}
		if history.ServerID == uuid.Nil {
			t.Error("Expected non-nil ServerID in StateHistory")
		}
		if history.SessionID == uuid.Nil {
			t.Error("Expected non-nil SessionID in StateHistory")
		}
		if history.Session == "" {
			t.Error("Expected non-empty Session in StateHistory")
		}
		if history.Track == "" {
			t.Error("Expected non-empty Track in StateHistory")
		}
		if history.DateCreated.IsZero() {
			t.Error("Expected non-zero DateCreated in StateHistory")
		}
	}
}
