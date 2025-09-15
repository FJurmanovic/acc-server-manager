package testdata

import (
	"acc-server-manager/local/model"
	"time"

	"github.com/google/uuid"
)

// StateHistoryTestData provides simple test data generators
type StateHistoryTestData struct {
	ServerID uuid.UUID
	BaseTime time.Time
}

// NewStateHistoryTestData creates a new test data generator
func NewStateHistoryTestData(serverID uuid.UUID) *StateHistoryTestData {
	return &StateHistoryTestData{
		ServerID: serverID,
		BaseTime: time.Now().UTC(),
	}
}

// CreateStateHistory creates a basic state history entry
func (td *StateHistoryTestData) CreateStateHistory(session model.TrackSession, track string, playerCount int, sessionID uuid.UUID) model.StateHistory {
	return model.StateHistory{
		ID:                     uuid.New(),
		ServerID:               td.ServerID,
		Session:                session,
		Track:                  track,
		PlayerCount:            playerCount,
		DateCreated:            td.BaseTime,
		SessionStart:           td.BaseTime,
		SessionDurationMinutes: 30,
		SessionID:              sessionID,
	}
}

// CreateMultipleEntries creates multiple state history entries for the same session
func (td *StateHistoryTestData) CreateMultipleEntries(session model.TrackSession, track string, playerCounts []int) []model.StateHistory {
	sessionID := uuid.New()
	var entries []model.StateHistory

	for i, count := range playerCounts {
		entry := model.StateHistory{
			ID:                     uuid.New(),
			ServerID:               td.ServerID,
			Session:                session,
			Track:                  track,
			PlayerCount:            count,
			DateCreated:            td.BaseTime.Add(time.Duration(i*5) * time.Minute),
			SessionStart:           td.BaseTime,
			SessionDurationMinutes: 30,
			SessionID:              sessionID,
		}
		entries = append(entries, entry)
	}

	return entries
}

// CreateBasicFilter creates a basic filter for testing
func CreateBasicFilter(serverID string) *model.StateHistoryFilter {
	return &model.StateHistoryFilter{
		ServerBasedFilter: model.ServerBasedFilter{
			ServerID: serverID,
		},
	}
}

// CreateFilterWithSession creates a filter with session type
func CreateFilterWithSession(serverID string, session model.TrackSession) *model.StateHistoryFilter {
	return &model.StateHistoryFilter{
		ServerBasedFilter: model.ServerBasedFilter{
			ServerID: serverID,
		},
		Session: session,
	}
}

// LogLines contains sample ACC server log lines for testing
var SampleLogLines = []string{
	"[2024-01-15 14:30:25.123] Session changed: NONE -> PRACTICE",
	"[2024-01-15 14:30:30.456] 1 client(s) online",
	"[2024-01-15 14:30:35.789] 3 client(s) online",
	"[2024-01-15 14:31:00.123] 5 client(s) online",
	"[2024-01-15 14:35:00.456] Session changed: PRACTICE -> QUALIFY",
	"[2024-01-15 14:35:05.789] 8 client(s) online",
	"[2024-01-15 14:40:00.123] Session changed: QUALIFY -> RACE",
	"[2024-01-15 14:40:05.456] 12 client(s) online",
	"[2024-01-15 14:45:00.789] 15 client(s) online",
	"[2024-01-15 14:50:00.123] Removing dead connection",
	"[2024-01-15 14:50:05.456] 14 client(s) online",
	"[2024-01-15 15:00:00.789] 0 client(s) online",
	"[2024-01-15 15:00:05.123] Session changed: RACE -> NONE",
}

// ExpectedSessionChanges represents the expected session changes from parsing the sample log lines
var ExpectedSessionChanges = []struct {
	From model.TrackSession
	To   model.TrackSession
}{
	{model.SessionUnknown, model.SessionPractice},
	{model.SessionPractice, model.SessionQualify},
	{model.SessionQualify, model.SessionRace},
	{model.SessionRace, model.SessionUnknown},
}

// ExpectedPlayerCounts represents the expected player counts from parsing the sample log lines
var ExpectedPlayerCounts = []int{1, 3, 5, 8, 12, 15, 14, 0}
