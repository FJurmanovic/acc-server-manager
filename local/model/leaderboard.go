package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Leaderboard struct {
	ID          uuid.UUID             `gorm:"type:uuid;primary_key;" json:"id"`
	ServerID    uuid.UUID             `gorm:"uniqueIndex;not null;type:uuid" json:"serverId"`
	FLPoints    int                   `gorm:"default:1" json:"flPoints"`
	FLColor     string                `gorm:"default:'#8b5cf6'" json:"flColor"`
	FLTextColor string                `gorm:"default:'#000000'" json:"flTextColor"`
	Drivers     []LeaderboardDriver   `gorm:"foreignKey:LeaderboardID;constraint:OnDelete:CASCADE" json:"drivers"`
	Races       []LeaderboardRace     `gorm:"foreignKey:LeaderboardID;constraint:OnDelete:CASCADE" json:"races"`
	PointRows   []LeaderboardPointRow `gorm:"foreignKey:LeaderboardID;constraint:OnDelete:CASCADE" json:"pointRows"`
}

type LeaderboardDriver struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	LeaderboardID uuid.UUID `gorm:"not null;type:uuid" json:"leaderboardId"`
	Name          string    `gorm:"not null" json:"name"`
	Initials      string    `json:"initials"`
	Color         string    `json:"color"`
	Position      int       `json:"position"`
}

type LeaderboardRace struct {
	ID                 uuid.UUID           `gorm:"type:uuid;primary_key;" json:"id"`
	LeaderboardID      uuid.UUID           `gorm:"not null;type:uuid" json:"leaderboardId"`
	Name               string              `gorm:"not null" json:"name"`
	Position           int                 `json:"position"`
	FastestLapInitials string              `json:"fastestLapInitials"`
	Results            []LeaderboardResult `gorm:"foreignKey:RaceID;constraint:OnDelete:CASCADE" json:"results"`
}

type LeaderboardResult struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	RaceID   uuid.UUID `gorm:"not null;type:uuid" json:"raceId"`
	DriverID uuid.UUID `gorm:"not null;type:uuid" json:"driverId"`
	Score    string    `json:"score"`
}

type LeaderboardPointRow struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	LeaderboardID uuid.UUID `gorm:"not null;type:uuid" json:"leaderboardId"`
	Label         string    `json:"label"`
	Points        int       `json:"points"`
	Color         string    `json:"color"`
	TextColor     string    `json:"textColor"`
	Priority      int       `json:"priority"`
}

// LeaderboardInput is the wire format matching the HTML's appData shape.
type LeaderboardInput struct {
	Drivers     []LeaderboardDriverInput   `json:"drivers"`
	PointsTable []LeaderboardPointRowInput `json:"pointsTable"`
	FLPoints    LeaderboardFLInput         `json:"flPoints"`
	Tracks      []LeaderboardTrackInput    `json:"tracks"`
}

type LeaderboardDriverInput struct {
	Name     string `json:"name"`
	Color    string `json:"color"`
	Initials string `json:"initials"`
}

type LeaderboardPointRowInput struct {
	Points    int    `json:"points"`
	Label     string `json:"label"`
	Color     string `json:"color"`
	TextColor string `json:"textColor"`
	Priority  int    `json:"priority"`
}

type LeaderboardFLInput struct {
	Points    int    `json:"points"`
	Label     string `json:"label"`
	Color     string `json:"color"`
	TextColor string `json:"textColor"`
	Priority  int    `json:"priority"`
}

type LeaderboardTrackInput struct {
	Name               string        `json:"name"`
	Results            []interface{} `json:"results"`
	FastestLapInitials string        `json:"fastestLapInitials"`
}

func (l *Leaderboard) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

func (d *LeaderboardDriver) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

func (r *LeaderboardRace) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (r *LeaderboardResult) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (p *LeaderboardPointRow) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
