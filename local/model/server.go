package model

import (
	"sync"
	"time"
)

// Server represents an ACC server instance
type Server struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Status      ServiceStatus `json:"status" gorm:"-"`
	IP          string `gorm:"not null" json:"-"`
	Port        int    `gorm:"not null" json:"-"`
	ConfigPath  string `gorm:"not null" json:"-"` // e.g. "/acc/servers/server1/"
	ServiceName string `gorm:"not null" json:"-"` // Windows service name
    State       ServerState `gorm:"-" json:"state"`
}

type PlayerState struct {
    CarID       int     // Car ID in broadcast packets
    DriverName  string  // Optional: pulled from registration packet
    TeamName    string
    CarModel    string
    CurrentLap  int
    LastLapTime int     // in milliseconds
    BestLapTime int     // in milliseconds
    Position    int
    ConnectedAt time.Time
    DisconnectedAt *time.Time
    IsConnected bool
}

type State struct {
    Session     string `json:"session"`
    SessionStart time.Time  `json:"sessionStart"`
    PlayerCount int `json:"playerCount"`
    // Players     map[int]*PlayerState
    // etc.
}

type ServerState struct {
    sync.RWMutex
    Session     string `json:"session"`
    SessionStart time.Time  `json:"sessionStart"`
    PlayerCount int `json:"playerCount"`
    Track       string `json:"track"`
    MaxConnections int `json:"maxConnections"`
    // Players     map[int]*PlayerState
    // etc.
}