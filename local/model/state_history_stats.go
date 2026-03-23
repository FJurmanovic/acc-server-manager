package model

import "github.com/google/uuid"

type SessionCount struct {
	Name  TrackSession `json:"name"`
	Count int          `json:"count"`
}

type DailyActivity struct {
	Date          string `json:"date"`
	SessionsCount int    `json:"sessionsCount"`
}

type PlayerCountPoint struct {
	Timestamp string  `json:"timestamp"`
	Count     float64 `json:"count"`
}

type StateHistoryStats struct {
	AveragePlayers      float64            `json:"averagePlayers"`
	PeakPlayers         int                `json:"peakPlayers"`
	TotalSessions       int                `json:"totalSessions"`
	TotalPlaytime       int                `json:"totalPlaytime" gorm:"-"`
	PlayerCountOverTime []PlayerCountPoint `json:"playerCountOverTime" gorm:"-"`
	SessionTypes        []SessionCount     `json:"sessionTypes" gorm:"-"`
	DailyActivity       []DailyActivity    `json:"dailyActivity" gorm:"-"`
	RecentSessions      []RecentSession    `json:"recentSessions" gorm:"-"`
}

type RecentSession struct {
	ID       uuid.UUID    `json:"id"`
	Date     string       `json:"date"`
	Type     TrackSession `json:"type"`
	Track    string       `json:"track"`
	Duration int          `json:"duration"`
	Players  int          `json:"players"`
}
