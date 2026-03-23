package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StateHistoryFilter struct {
	ServerBasedFilter
	DateRangeFilter

	Session    TrackSession `query:"session"`
	MinPlayers *int         `query:"min_players"`
	MaxPlayers *int         `query:"max_players"`
}

func (f *StateHistoryFilter) ApplyFilter(query *gorm.DB) *gorm.DB {
	if f.ServerID != "" {
		if serverUUID, err := uuid.Parse(f.ServerID); err == nil {
			query = query.Where("server_id = ?", serverUUID)
		}
	}

	timeZero := time.Time{}
	if f.StartDate != timeZero {
		query = query.Where("date_created >= ?", f.StartDate)
	}
	if f.EndDate != timeZero {
		query = query.Where("date_created <= ?", f.EndDate)
	}

	if f.Session != "" {
		query = query.Where("session = ?", f.Session)
	}

	if f.MinPlayers != nil {
		query = query.Where("player_count >= ?", *f.MinPlayers)
	}
	if f.MaxPlayers != nil {
		query = query.Where("player_count <= ?", *f.MaxPlayers)
	}

	return query
}

type TrackSession string

const (
	SessionPractice TrackSession = "P"
	SessionQualify  TrackSession = "Q"
	SessionRace     TrackSession = "R"
	SessionUnknown  TrackSession = "U"
)

func (i *TrackSession) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err == nil {
		*i = ToTrackSession(str)
		return nil
	}

	return fmt.Errorf("invalid TrackSession value")
}

func (i TrackSession) Humanize() string {
	switch i {
	case SessionPractice:
		return "Practice"
	case SessionQualify:
		return "Qualifying"
	case SessionRace:
		return "Race"
	default:
		return "Unknown"
	}
}

func ToTrackSession(i string) TrackSession {
	sessionAbrv := strings.ToUpper(i[:1])
	switch sessionAbrv {
	case "P":
		return SessionPractice
	case "Q":
		return SessionQualify
	case "R":
		return SessionRace
	default:
		return SessionUnknown
	}
}

func (i TrackSession) ToString() string {
	return string(i)
}

type StateHistory struct {
	ID                     uuid.UUID    `gorm:"type:uuid;primary_key;" json:"id"`
	ServerID               uuid.UUID    `json:"serverId" gorm:"not null;type:uuid"`
	Session                TrackSession `json:"session"`
	Track                  string       `json:"track"`
	PlayerCount            int          `json:"playerCount"`
	DateCreated            time.Time    `json:"dateCreated"`
	SessionStart           time.Time    `json:"sessionStart"`
	SessionDurationMinutes int          `json:"sessionDurationMinutes"`
	SessionID              uuid.UUID    `json:"sessionId" gorm:"not null;type:uuid"`
}

func (sh *StateHistory) BeforeCreate(tx *gorm.DB) error {
	if sh.ID == uuid.Nil {
		sh.ID = uuid.New()
	}
	if sh.SessionID == uuid.Nil {
		sh.SessionID = uuid.New()
	}
	if sh.DateCreated.IsZero() {
		sh.DateCreated = time.Now().UTC()
	}
	return nil
}
