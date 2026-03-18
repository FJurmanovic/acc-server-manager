package model

import (
	"encoding/json"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"strconv"
)

// Score is a custom type that accepts both numbers and strings in JSON.
type Score string

func (s *Score) UnmarshalJSON(data []byte) error {
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*s = Score(n.String())
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = Score(str)
	return nil
}

func (s Score) MarshalJSON() ([]byte, error) {
	if n, err := strconv.Atoi(string(s)); err == nil {
		return json.Marshal(n)
	}
	return json.Marshal(string(s))
}

type Leaderboard struct {
	ID          uuid.UUID             `gorm:"type:uuid;primary_key;" json:"id"`
	ServerID    uuid.UUID             `gorm:"uniqueIndex;not null;type:uuid" json:"serverId"`
	FLPoints    int                   `gorm:"default:1" json:"flPoints"`
	FLColor     string                `gorm:"default:'#8b5cf6'" json:"flColor"`
	FLTextColor string                `gorm:"default:'#000000'" json:"flTextColor"`
	Drivers     []LeaderboardDriver   `gorm:"foreignKey:LeaderboardID;constraint:OnDelete:CASCADE" json:"drivers"`
	Races       []LeaderboardRace     `gorm:"foreignKey:LeaderboardID;constraint:OnDelete:CASCADE" json:"tracks"`
	PointRows   []LeaderboardPointRow `gorm:"foreignKey:LeaderboardID;constraint:OnDelete:CASCADE" json:"pointsTable"`
}

type LeaderboardDriver struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	LeaderboardID uuid.UUID `gorm:"not null;type:uuid" json:"-"`
	Name          string    `gorm:"not null" json:"name"`
	Initials      string    `json:"initials"`
	Color         string    `json:"color"`
	Position      int       `json:"-"`
}

type LeaderboardRace struct {
	ID                 uuid.UUID           `gorm:"type:uuid;primary_key;" json:"id"`
	LeaderboardID      uuid.UUID           `gorm:"not null;type:uuid" json:"-"`
	Name               string              `gorm:"not null" json:"name"`
	Position           int                 `json:"-"`
	FastestLapDriverID *uuid.UUID          `gorm:"type:uuid" json:"fastestLapDriverId"`
	Results            []LeaderboardResult `gorm:"foreignKey:RaceID;constraint:OnDelete:CASCADE" json:"results"`
}

type LeaderboardResult struct {
	ID       uuid.UUID `gorm:"type:uuid;primary_key;" json:"-"`
	RaceID   uuid.UUID `gorm:"not null;type:uuid" json:"-"`
	DriverID uuid.UUID `gorm:"not null;type:uuid" json:"driverId"`
	Score    Score     `json:"score"`
}

type LeaderboardPointRow struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	LeaderboardID uuid.UUID `gorm:"not null;type:uuid" json:"-"`
	Label         string    `json:"label"`
	Points        int       `json:"points"`
	Color         string    `json:"color"`
	TextColor     string    `json:"textColor"`
	Priority      int       `json:"priority"`
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
