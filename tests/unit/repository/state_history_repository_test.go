package repository

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/tests"
	"acc-server-manager/tests/testdata"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStateHistoryRepository_Insert_Success(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)
	history := testData.CreateStateHistory(model.SessionPractice, "spa", 5, uuid.New())

	err := repo.Insert(ctx, &history)
	tests.AssertNoError(t, err)

	tests.AssertNotNil(t, history.ID)
	if history.ID == uuid.Nil {
		t.Error("Expected non-nil ID after insert")
	}
}

func TestStateHistoryRepository_GetAll_Success(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)

	playerCounts := []int{0, 5, 10, 15, 10, 5, 0}
	entries := testData.CreateMultipleEntries(model.SessionPractice, "spa", playerCounts)

	for _, entry := range entries {
		err := repo.Insert(ctx, &entry)
		tests.AssertNoError(t, err)
	}

	filter := testdata.CreateBasicFilter(helper.TestData.ServerID.String())
	result, err := repo.GetAll(ctx, filter)

	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, result)
	tests.AssertEqual(t, len(entries), len(*result))
}

func TestStateHistoryRepository_GetAll_WithFilter(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)

	practiceHistory := testData.CreateStateHistory(model.SessionPractice, "spa", 5, uuid.New())
	raceHistory := testData.CreateStateHistory(model.SessionRace, "spa", 15, uuid.New())

	err := repo.Insert(ctx, &practiceHistory)
	tests.AssertNoError(t, err)
	err = repo.Insert(ctx, &raceHistory)
	tests.AssertNoError(t, err)

	filter := testdata.CreateFilterWithSession(helper.TestData.ServerID.String(), model.SessionRace)
	result, err := repo.GetAll(ctx, filter)

	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, result)
	tests.AssertEqual(t, 1, len(*result))
	tests.AssertEqual(t, model.SessionRace, (*result)[0].Session)
	tests.AssertEqual(t, 15, (*result)[0].PlayerCount)
}

func TestStateHistoryRepository_GetLastSessionID_Success(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)

	sessionID1 := uuid.New()
	sessionID2 := uuid.New()

	history1 := testData.CreateStateHistory(model.SessionPractice, "spa", 5, sessionID1)
	history2 := testData.CreateStateHistory(model.SessionRace, "spa", 10, sessionID2)

	err := repo.Insert(ctx, &history1)
	tests.AssertNoError(t, err)

	time.Sleep(1 * time.Millisecond)

	err = repo.Insert(ctx, &history2)
	tests.AssertNoError(t, err)
	lastSessionID, err := repo.GetLastSessionID(ctx, helper.TestData.ServerID)
	tests.AssertNoError(t, err)

	if lastSessionID == uuid.Nil {
		t.Fatal("Expected non-nil UUID for last session ID")
	}
}

func TestStateHistoryRepository_GetLastSessionID_NoData(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()

	lastSessionID, err := repo.GetLastSessionID(ctx, helper.TestData.ServerID)
	tests.AssertNoError(t, err)
	tests.AssertEqual(t, uuid.Nil, lastSessionID)
}

func TestStateHistoryRepository_GetSummaryStats_Success(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)

	sessionID1 := uuid.New()
	sessionID2 := uuid.New()

	practiceEntries := testData.CreateMultipleEntries(model.SessionPractice, "spa", []int{5, 10, 15})
	for i := range practiceEntries {
		practiceEntries[i].SessionID = sessionID1
		err := repo.Insert(ctx, &practiceEntries[i])
		tests.AssertNoError(t, err)
	}

	raceEntries := testData.CreateMultipleEntries(model.SessionRace, "spa", []int{20, 25, 30})
	for i := range raceEntries {
		raceEntries[i].SessionID = sessionID2
		err := repo.Insert(ctx, &raceEntries[i])
		tests.AssertNoError(t, err)
	}

	filter := testdata.CreateBasicFilter(helper.TestData.ServerID.String())
	stats, err := repo.GetSummaryStats(ctx, filter)

	tests.AssertNoError(t, err)

	tests.AssertEqual(t, 30, stats.PeakPlayers)
	tests.AssertEqual(t, 2, stats.TotalSessions)

	expectedAverage := float64(5+10+15+20+25+30) / 6.0
	if stats.AveragePlayers != expectedAverage {
		t.Errorf("Expected average players %.1f, got %.1f", expectedAverage, stats.AveragePlayers)
	}
}

func TestStateHistoryRepository_GetSummaryStats_NoData(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()

	filter := testdata.CreateBasicFilter(helper.TestData.ServerID.String())
	stats, err := repo.GetSummaryStats(ctx, filter)

	tests.AssertNoError(t, err)

	tests.AssertEqual(t, 0, stats.PeakPlayers)
	tests.AssertEqual(t, 0.0, stats.AveragePlayers)
	tests.AssertEqual(t, 0, stats.TotalSessions)
}

func TestStateHistoryRepository_GetTotalPlaytime_Success(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()
	sessionID := uuid.New()

	baseTime := time.Now().UTC()

	entries := []model.StateHistory{
		{
			ID:                     uuid.New(),
			ServerID:               helper.TestData.ServerID,
			Session:                model.SessionPractice,
			Track:                  "spa",
			PlayerCount:            5,
			DateCreated:            baseTime,
			SessionStart:           baseTime,
			SessionDurationMinutes: 30,
			SessionID:              sessionID,
		},
		{
			ID:                     uuid.New(),
			ServerID:               helper.TestData.ServerID,
			Session:                model.SessionPractice,
			Track:                  "spa",
			PlayerCount:            10,
			DateCreated:            baseTime.Add(30 * time.Minute),
			SessionStart:           baseTime,
			SessionDurationMinutes: 30,
			SessionID:              sessionID,
		},
		{
			ID:                     uuid.New(),
			ServerID:               helper.TestData.ServerID,
			Session:                model.SessionPractice,
			Track:                  "spa",
			PlayerCount:            8,
			DateCreated:            baseTime.Add(60 * time.Minute),
			SessionStart:           baseTime,
			SessionDurationMinutes: 30,
			SessionID:              sessionID,
		},
	}

	for _, entry := range entries {
		err := repo.Insert(ctx, &entry)
		tests.AssertNoError(t, err)
	}

	filter := &model.StateHistoryFilter{
		ServerBasedFilter: model.ServerBasedFilter{
			ServerID: helper.TestData.ServerID.String(),
		},
		DateRangeFilter: model.DateRangeFilter{
			StartDate: baseTime.Add(-1 * time.Hour),
			EndDate:   baseTime.Add(2 * time.Hour),
		},
	}

	playtime, err := repo.GetTotalPlaytime(ctx, filter)
	tests.AssertNoError(t, err)
	if playtime <= 0 {
		t.Error("Expected positive playtime for session with multiple entries")
	}
}

func TestStateHistoryRepository_ConcurrentOperations(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	initialHistory := testData.CreateStateHistory(model.SessionPractice, "spa", 5, uuid.New())
	err := repo.Insert(ctx, &initialHistory)
	if err != nil {
		t.Fatalf("Failed to insert initial record: %v", err)
	}

	done := make(chan bool, 3)

	go func() {
		defer func() {
			done <- true
		}()
		history := testData.CreateStateHistory(model.SessionPractice, "spa", 5, uuid.New())
		err := repo.Insert(ctx, &history)
		if err != nil {
			t.Logf("Insert error: %v", err)
			return
		}
	}()

	go func() {
		defer func() {
			done <- true
		}()
		filter := testdata.CreateBasicFilter(helper.TestData.ServerID.String())
		_, err := repo.GetAll(ctx, filter)
		if err != nil {
			t.Logf("GetAll error: %v", err)
			return
		}
	}()

	go func() {
		defer func() {
			done <- true
		}()
		_, err := repo.GetLastSessionID(ctx, helper.TestData.ServerID)
		if err != nil {
			t.Logf("GetLastSessionID error: %v", err)
			return
		}
	}()

	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestStateHistoryRepository_FilterEdgeCases(t *testing.T) {
	tests.SetTestEnv()
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	if !helper.DB.Migrator().HasTable(&model.StateHistory{}) {
		err := helper.DB.Migrator().CreateTable(&model.StateHistory{})
		if err != nil {
			t.Fatalf("Failed to create state_histories table: %v", err)
		}
	}

	repo := repository.NewStateHistoryRepository(helper.DB)
	ctx := helper.CreateContext()

	testData := testdata.NewStateHistoryTestData(helper.TestData.ServerID)
	history := testData.CreateStateHistory(model.SessionPractice, "spa", 5, uuid.New())
	err := repo.Insert(ctx, &history)
	tests.AssertNoError(t, err)

	serverFilter := &model.StateHistoryFilter{
		ServerBasedFilter: model.ServerBasedFilter{
			ServerID: helper.TestData.ServerID.String(),
		},
	}
	result, err := repo.GetAll(ctx, serverFilter)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, result)

	invalidFilter := &model.StateHistoryFilter{
		ServerBasedFilter: model.ServerBasedFilter{
			ServerID: "invalid-uuid",
		},
	}
	_, err = repo.GetSummaryStats(ctx, invalidFilter)
	if err == nil {
		t.Error("Expected error for invalid server ID in GetSummaryStats")
	}
}
