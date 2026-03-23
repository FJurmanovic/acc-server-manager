package mocks

import (
	"acc-server-manager/local/model"
	"context"
	"errors"

	"github.com/google/uuid"
)

type MockStateHistoryRepository struct {
	stateHistories   []model.StateHistory
	shouldFailGet    bool
	shouldFailInsert bool
}

func NewMockStateHistoryRepository() *MockStateHistoryRepository {
	return &MockStateHistoryRepository{
		stateHistories: make([]model.StateHistory, 0),
	}
}

func (m *MockStateHistoryRepository) GetAll(ctx context.Context, filter *model.StateHistoryFilter) (*[]model.StateHistory, error) {
	if m.shouldFailGet {
		return nil, errors.New("failed to get state history")
	}

	var filtered []model.StateHistory
	for _, sh := range m.stateHistories {
		if m.matchesFilter(sh, filter) {
			filtered = append(filtered, sh)
		}
	}

	return &filtered, nil
}

func (m *MockStateHistoryRepository) Insert(ctx context.Context, stateHistory *model.StateHistory) error {
	if m.shouldFailInsert {
		return errors.New("failed to insert state history")
	}

	if stateHistory.ID == uuid.Nil {
		stateHistory.ID = uuid.New()
	}
	if stateHistory.SessionID == uuid.Nil {
		stateHistory.SessionID = uuid.New()
	}

	m.stateHistories = append(m.stateHistories, *stateHistory)
	return nil
}

func (m *MockStateHistoryRepository) GetLastSessionID(ctx context.Context, serverID uuid.UUID) (uuid.UUID, error) {
	for i := len(m.stateHistories) - 1; i >= 0; i-- {
		if m.stateHistories[i].ServerID == serverID {
			return m.stateHistories[i].SessionID, nil
		}
	}
	return uuid.Nil, nil
}

func (m *MockStateHistoryRepository) matchesFilter(sh model.StateHistory, filter *model.StateHistoryFilter) bool {
	if filter == nil {
		return true
	}

	if filter.ServerID != "" {
		serverUUID, err := uuid.Parse(filter.ServerID)
		if err != nil || sh.ServerID != serverUUID {
			return false
		}
	}

	if filter.Session != "" && sh.Session != filter.Session {
		return false
	}

	if filter.MinPlayers != nil && sh.PlayerCount < *filter.MinPlayers {
		return false
	}

	if filter.MaxPlayers != nil && sh.PlayerCount > *filter.MaxPlayers {
		return false
	}

	return true
}

func (m *MockStateHistoryRepository) SetShouldFailGet(shouldFail bool) {
	m.shouldFailGet = shouldFail
}

func (m *MockStateHistoryRepository) SetShouldFailInsert(shouldFail bool) {
	m.shouldFailInsert = shouldFail
}

func (m *MockStateHistoryRepository) AddStateHistory(stateHistory model.StateHistory) {
	if stateHistory.ID == uuid.Nil {
		stateHistory.ID = uuid.New()
	}
	if stateHistory.SessionID == uuid.Nil {
		stateHistory.SessionID = uuid.New()
	}
	m.stateHistories = append(m.stateHistories, stateHistory)
}

func (m *MockStateHistoryRepository) GetCount() int {
	return len(m.stateHistories)
}

func (m *MockStateHistoryRepository) Clear() {
	m.stateHistories = make([]model.StateHistory, 0)
}

func (m *MockStateHistoryRepository) GetSummaryStats(ctx context.Context, filter *model.StateHistoryFilter) (model.StateHistoryStats, error) {
	var stats model.StateHistoryStats
	var filteredEntries []model.StateHistory

	for _, entry := range m.stateHistories {
		if m.matchesFilter(entry, filter) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	if len(filteredEntries) == 0 {
		return stats, nil
	}

	sessionMap := make(map[string]bool)
	totalPlayers := 0

	for _, entry := range filteredEntries {
		if entry.PlayerCount > stats.PeakPlayers {
			stats.PeakPlayers = entry.PlayerCount
		}
		totalPlayers += entry.PlayerCount
		sessionMap[entry.SessionID.String()] = true
	}

	stats.TotalSessions = len(sessionMap)
	if len(filteredEntries) > 0 {
		stats.AveragePlayers = float64(totalPlayers) / float64(len(filteredEntries))
	}

	return stats, nil
}

func (m *MockStateHistoryRepository) GetTotalPlaytime(ctx context.Context, filter *model.StateHistoryFilter) (int, error) {
	var filteredEntries []model.StateHistory

	for _, entry := range m.stateHistories {
		if m.matchesFilter(entry, filter) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	if len(filteredEntries) == 0 {
		return 0, nil
	}

	sessionMap := make(map[string][]model.StateHistory)
	for _, entry := range filteredEntries {
		sessionID := entry.SessionID.String()
		sessionMap[sessionID] = append(sessionMap[sessionID], entry)
	}

	totalMinutes := 0
	for _, sessionEntries := range sessionMap {
		if len(sessionEntries) > 1 {
			minTime := sessionEntries[0].DateCreated
			maxTime := sessionEntries[0].DateCreated
			hasPlayers := false

			for _, entry := range sessionEntries {
				if entry.DateCreated.Before(minTime) {
					minTime = entry.DateCreated
				}
				if entry.DateCreated.After(maxTime) {
					maxTime = entry.DateCreated
				}
				if entry.PlayerCount > 0 {
					hasPlayers = true
				}
			}

			if hasPlayers {
				duration := maxTime.Sub(minTime)
				totalMinutes += int(duration.Minutes())
			}
		}
	}

	return totalMinutes, nil
}

func (m *MockStateHistoryRepository) GetPlayerCountOverTime(ctx context.Context, filter *model.StateHistoryFilter) ([]model.PlayerCountPoint, error) {
	var points []model.PlayerCountPoint
	var filteredEntries []model.StateHistory

	for _, entry := range m.stateHistories {
		if m.matchesFilter(entry, filter) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	hourMap := make(map[string][]int)
	for _, entry := range filteredEntries {
		hourKey := entry.DateCreated.Format("2006-01-02 15")
		hourMap[hourKey] = append(hourMap[hourKey], entry.PlayerCount)
	}

	for hourKey, counts := range hourMap {
		total := 0
		for _, count := range counts {
			total += count
		}
		avg := total / len(counts)

		points = append(points, model.PlayerCountPoint{
			Timestamp: hourKey,
			Count:     float64(avg),
		})
	}

	return points, nil
}

func (m *MockStateHistoryRepository) GetSessionTypes(ctx context.Context, filter *model.StateHistoryFilter) ([]model.SessionCount, error) {
	var sessionTypes []model.SessionCount
	var filteredEntries []model.StateHistory

	for _, entry := range m.stateHistories {
		if m.matchesFilter(entry, filter) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	sessionMap := make(map[model.TrackSession]map[string]bool)
	for _, entry := range filteredEntries {
		if sessionMap[entry.Session] == nil {
			sessionMap[entry.Session] = make(map[string]bool)
		}
		sessionMap[entry.Session][entry.SessionID.String()] = true
	}

	for sessionType, sessions := range sessionMap {
		sessionTypes = append(sessionTypes, model.SessionCount{
			Name:  sessionType,
			Count: len(sessions),
		})
	}

	return sessionTypes, nil
}

func (m *MockStateHistoryRepository) GetDailyActivity(ctx context.Context, filter *model.StateHistoryFilter) ([]model.DailyActivity, error) {
	var dailyActivity []model.DailyActivity
	var filteredEntries []model.StateHistory

	for _, entry := range m.stateHistories {
		if m.matchesFilter(entry, filter) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	dayMap := make(map[string]map[string]bool)
	for _, entry := range filteredEntries {
		dateKey := entry.DateCreated.Format("2006-01-02")
		if dayMap[dateKey] == nil {
			dayMap[dateKey] = make(map[string]bool)
		}
		dayMap[dateKey][entry.SessionID.String()] = true
	}

	for date, sessions := range dayMap {
		dailyActivity = append(dailyActivity, model.DailyActivity{
			Date:          date,
			SessionsCount: len(sessions),
		})
	}

	return dailyActivity, nil
}

func (m *MockStateHistoryRepository) GetRecentSessions(ctx context.Context, filter *model.StateHistoryFilter) ([]model.RecentSession, error) {
	var recentSessions []model.RecentSession
	var filteredEntries []model.StateHistory

	for _, entry := range m.stateHistories {
		if m.matchesFilter(entry, filter) {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	sessionMap := make(map[string][]model.StateHistory)
	for _, entry := range filteredEntries {
		sessionID := entry.SessionID.String()
		sessionMap[sessionID] = append(sessionMap[sessionID], entry)
	}

	count := 0
	for _, entries := range sessionMap {
		if count >= 10 {
			break
		}

		if len(entries) > 0 {
			minDate := entries[0].DateCreated
			maxDate := entries[0].DateCreated
			maxPlayers := 0

			for _, entry := range entries {
				if entry.DateCreated.Before(minDate) {
					minDate = entry.DateCreated
				}
				if entry.DateCreated.After(maxDate) {
					maxDate = entry.DateCreated
				}
				if entry.PlayerCount > maxPlayers {
					maxPlayers = entry.PlayerCount
				}
			}

			if maxPlayers > 0 {
				duration := int(maxDate.Sub(minDate).Minutes())
				recentSessions = append(recentSessions, model.RecentSession{
					ID:       entries[0].SessionID,
					Date:     minDate.Format("2006-01-02 15:04:05"),
					Type:     entries[0].Session,
					Track:    entries[0].Track,
					Players:  maxPlayers,
					Duration: duration,
				})
				count++
			}
		}
	}

	return recentSessions, nil
}
