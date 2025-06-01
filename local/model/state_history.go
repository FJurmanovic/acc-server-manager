package model

import (
	"time"

	"gorm.io/gorm"
)

// StateHistoryFilter combines common filter capabilities
type StateHistoryFilter struct {
	ServerBasedFilter // Adds server ID from path parameter
	DateRangeFilter   // Adds date range filtering
	
	// Additional fields specific to state history
	Session     string `query:"session"`
	MinPlayers  *int   `query:"min_players"`
	MaxPlayers  *int   `query:"max_players"`
}

// ApplyFilter implements the Filterable interface
func (f *StateHistoryFilter) ApplyFilter(query *gorm.DB) *gorm.DB {
	// Apply server filter
	if f.ServerID != 0 {
		query = query.Where("server_id = ?", f.ServerID)
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
	ID          uint      `gorm:"primaryKey" json:"id"`
	ServerID    uint      `json:"serverId" gorm:"not null"`
	Session     string    `json:"session"`
	Track       string    `json:"track"`
	PlayerCount int       `json:"playerCount"`
	DateCreated time.Time `json:"dateCreated"`
	SessionStart time.Time `json:"sessionStart"`
	SessionDurationMinutes int `json:"sessionDurationMinutes"`
	SessionID   uint      `json:"sessionId" gorm:"not null;default:0"` // Unique identifier for each session/event
}