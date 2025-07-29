package model

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	BaseServerPath    = "servers"
	ServiceNamePrefix = "ACC-Server"
)

// Server represents an ACC server instance
type ServerAPI struct {
	Name        string        `json:"name"`
	Status      ServiceStatus `json:"status"`
	State       *ServerState  `json:"state"`
	PlayerCount int           `json:"playerCount"`
	Track       string        `json:"track"`
}

func (s *Server) ToServerAPI() *ServerAPI {
	return &ServerAPI{
		Name:        s.Name,
		Status:      s.Status,
		State:       s.State,
		PlayerCount: s.State.PlayerCount,
		Track:       s.State.Track,
	}
}

// Server represents an ACC server instance
type Server struct {
	ID           uuid.UUID     `gorm:"type:uuid;primary_key;" json:"id"`
	Name         string        `gorm:"not null" json:"name"`
	Status       ServiceStatus `json:"status" gorm:"-"`
	IP           string        `gorm:"not null" json:"-"`
	Port         int           `gorm:"not null" json:"-"`
	Path         string        `gorm:"not null" json:"path"`        // e.g. "/acc/servers/server1/"
	ServiceName  string        `gorm:"not null" json:"serviceName"` // Windows service name
	State        *ServerState  `gorm:"-" json:"state"`
	DateCreated  time.Time     `json:"dateCreated"`
	FromSteamCMD bool          `gorm:"not null; default:true" json:"-"`
}

type PlayerState struct {
	CarID          int    // Car ID in broadcast packets
	DriverName     string // Optional: pulled from registration packet
	TeamName       string
	CarModel       string
	CurrentLap     int
	LastLapTime    int // in milliseconds
	BestLapTime    int // in milliseconds
	Position       int
	ConnectedAt    time.Time
	DisconnectedAt *time.Time
	IsConnected    bool
}

type State struct {
	Session      string    `json:"session"`
	SessionStart time.Time `json:"sessionStart"`
	PlayerCount  int       `json:"playerCount"`
	// Players     map[int]*PlayerState
	// etc.
}

type ServerState struct {
	sync.RWMutex
	Session                string    `json:"session"`
	SessionStart           time.Time `json:"sessionStart"`
	PlayerCount            int       `json:"playerCount"`
	Track                  string    `json:"track"`
	MaxConnections         int       `json:"maxConnections"`
	SessionDurationMinutes int       `json:"sessionDurationMinutes"`
	// Players     map[int]*PlayerState
	// etc.
}

// ServerFilter defines filtering options for Server queries
type ServerFilter struct {
	BaseFilter
	ServerBasedFilter
	Name        string `query:"name"`
	ServiceName string `query:"service_name"`
	Status      string `query:"status"`
}

// ApplyFilter implements the Filterable interface
func (f *ServerFilter) ApplyFilter(query *gorm.DB) *gorm.DB {
	// Apply server filter
	if f.ServerID != "" {
		if serverUUID, err := uuid.Parse(f.ServerID); err == nil {
			query = query.Where("id = ?", serverUUID)
		}
	}

	return query
}

// BeforeCreate is a GORM hook that runs before creating a new server
func (s *Server) BeforeCreate(tx *gorm.DB) error {
	if s.Name == "" {
		return errors.New("server name is required")
	}

	// Generate UUID if not set
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	// Generate service name and config path if not set
	if s.ServiceName == "" {
		s.ServiceName = s.GenerateServiceName()
	}
	if s.Path == "" {
		s.Path = s.GenerateServerPath(BaseServerPath)
	}

	// Set creation date if not set
	if s.DateCreated.IsZero() {
		s.DateCreated = time.Now().UTC()
	}

	return nil
}

// GenerateServiceName creates a unique service name based on the server name
func (s *Server) GenerateServiceName() string {
	// If ID is set, use it
	if s.ID != uuid.Nil {
		return fmt.Sprintf("%s-%s", ServiceNamePrefix, s.ID.String()[:8])
	}
	// Otherwise use a timestamp-based unique identifier
	return fmt.Sprintf("%s-%d", ServiceNamePrefix, time.Now().UnixNano())
}

// GenerateServerPath creates the config path based on the service name
func (s *Server) GenerateServerPath(steamCMDPath string) string {
	// Ensure service name is set
	if s.ServiceName == "" {
		s.ServiceName = s.GenerateServiceName()
	}
	if steamCMDPath == "" {
		steamCMDPath = BaseServerPath
	}
	return filepath.Join(steamCMDPath, "servers", s.ServiceName)
}

func (s *Server) GetServerPath() string {
	if !s.FromSteamCMD {
		return s.Path
	}
	return filepath.Join(s.Path, "server")
}

func (s *Server) GetConfigPath() string {
	return filepath.Join(s.GetServerPath(), "cfg")
}

func (s *Server) GetLogPath() string {
	if !s.FromSteamCMD {
		return s.Path
	}
	return filepath.Join(s.GetServerPath(), "log")
}

func (s *Server) Validate() error {
	if s.Name == "" {
		return errors.New("server name is required")
	}
	return nil
}
