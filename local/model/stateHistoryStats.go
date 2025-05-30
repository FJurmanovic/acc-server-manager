package model

import "time"

type SessionCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type DailyActivity struct {
	Date          time.Time `json:"date"`
	SessionsCount int       `json:"sessionsCount"`
}

type PlayerCountPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

type StateHistoryStats struct {
	AveragePlayers      float64           `json:"averagePlayers"`
	PeakPlayers         int               `json:"peakPlayers"`
	TotalSessions       int               `json:"totalSessions"`
	TotalPlaytime       int               `json:"totalPlaytime"` // in minutes
	PlayerCountOverTime []PlayerCountPoint `json:"playerCountOverTime"`
	SessionTypes        []SessionCount     `json:"sessionTypes"`
	DailyActivity       []DailyActivity    `json:"dailyActivity"`
	RecentSessions      []RecentSession    `json:"recentSessions"`
} 

type RecentSession struct {
	ID       uint      `json:"id"`
	Date     time.Time `json:"date"`
	Type     string    `json:"type"`
	Track    string    `json:"track"`
	Duration int       `json:"duration"`
	Players  int       `json:"players"`
}