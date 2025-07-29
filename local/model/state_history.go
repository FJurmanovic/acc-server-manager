package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StateHistoryFilter combines common filter capabilities
type StateHistoryFilter struct {
	ServerBasedFilter // Adds server ID from path parameter
	DateRangeFilter   // Adds date range filtering

	// Additional fields specific to state history
	Session    string `query:"session"`
	MinPlayers *int   `query:"min_players"`
	MaxPlayers *int   `query:"max_players"`
}

// ApplyFilter implements the Filterable interface
func (f *StateHistoryFilter) ApplyFilter(query *gorm.DB) *gorm.DB {
	// Apply server filter
	if f.ServerID != "" {
		if serverUUID, err := uuid.Parse(f.ServerID); err == nil {
			query = query.Where("server_id = ?", serverUUID)
		}
	}

	// Apply date range filter if set
	timeZero := time.Time{}
	if f.StartDate != timeZero {
		query = query.Where("date_created >= ?", f.StartDate)
	}
	if f.EndDate != timeZero {
		query = query.Where("date_created <= ?", f.EndDate)
	}

	// Apply session filter if set
	if f.Session != "" {
		query = query.Where("session = ?", f.Session)
	}

	// Apply player count filters if set
	if f.MinPlayers != nil {
		query = query.Where("player_count >= ?", *f.MinPlayers)
	}
	if f.MaxPlayers != nil {
		query = query.Where("player_count <= ?", *f.MaxPlayers)
	}

	return query
}

type StateHistory struct {
	ID                     uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	ServerID               uuid.UUID `json:"serverId" gorm:"not null;type:uuid"`
	Session                string    `json:"session"`
	Track                  string    `json:"track"`
	PlayerCount            int       `json:"playerCount"`
	DateCreated            time.Time `json:"dateCreated"`
	SessionStart           time.Time `json:"sessionStart"`
	SessionDurationMinutes int       `json:"sessionDurationMinutes"`
	SessionID              uuid.UUID `json:"sessionId" gorm:"not null;type:uuid"` // Unique identifier for each session/event
}

// BeforeCreate is a GORM hook that runs before creating new state history entries
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
