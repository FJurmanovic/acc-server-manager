package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActionType string

const (
	ActionConfigUpdate      ActionType = "config_update"
	ActionServerStart       ActionType = "server_start"
	ActionServerStop        ActionType = "server_stop"
	ActionServerRestart     ActionType = "server_restart"
	ActionLeaderboardUpdate ActionType = "leaderboard_update"
)

type ActivityLog struct {
	ID        uuid.UUID  `json:"id"        gorm:"type:uuid;primary_key;"`
	ServerID  uuid.UUID  `json:"serverId"  gorm:"not null;type:uuid;index"`
	Server    *Server    `json:"server,omitempty" gorm:"foreignKey:ServerID"`
	UserID    string     `json:"userId"    gorm:"not null"`
	Username  string     `json:"username"  gorm:"not null"`
	Action    ActionType `json:"action"    gorm:"not null"`
	Details   string     `json:"details"   gorm:"type:text"`
	CreatedAt time.Time  `json:"createdAt"`
}

func (a *ActivityLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	return nil
}

type ActivityLogFilter struct {
	BaseFilter
	DateRangeFilter
	UserID   string     `query:"user_id"`
	Action   ActionType `query:"action"`
	ServerID string     `query:"server_id"`
}

func (f *ActivityLogFilter) ApplyFilter(query *gorm.DB) *gorm.DB {
	if f.ServerID != "" {
		if serverUUID, err := uuid.Parse(f.ServerID); err == nil {
			query = query.Where("server_id = ?", serverUUID)
		}
	}
	if f.UserID != "" {
		query = query.Where("user_id = ?", f.UserID)
	}
	if f.Action != "" {
		query = query.Where("action = ?", f.Action)
	}
	timeZero := time.Time{}
	if f.StartDate != timeZero {
		query = query.Where("created_at >= ?", f.StartDate)
	}
	if f.EndDate != timeZero {
		query = query.Where("created_at <= ?", f.EndDate)
	}
	return query
}

func (f *ActivityLogFilter) GetSorting() (field string, desc bool) {
	if f.SortBy == "" {
		return "created_at", true
	}
	return f.SortBy, f.SortDesc
}

type LastActivityInfo struct {
	UpdatedAt time.Time  `json:"updatedAt"`
	Username  string     `json:"username"`
	Action    ActionType `json:"action"`
}
