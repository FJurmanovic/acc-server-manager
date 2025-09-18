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

type Server struct {
	ID           uuid.UUID     `gorm:"type:uuid;primary_key;" json:"id"`
	Name         string        `gorm:"not null" json:"name"`
	Status       ServiceStatus `json:"status" gorm:"-"`
	IP           string        `gorm:"not null" json:"-"`
	Port         int           `gorm:"not null" json:"-"`
	Path         string        `gorm:"not null" json:"path"`
	ServiceName  string        `gorm:"not null" json:"serviceName"`
	State        *ServerState  `gorm:"-" json:"state"`
	DateCreated  time.Time     `json:"dateCreated"`
	FromSteamCMD bool          `gorm:"not null; default:true" json:"-"`
}

type PlayerState struct {
	CarID          int
	DriverName     string
	TeamName       string
	CarModel       string
	CurrentLap     int
	LastLapTime    int
	BestLapTime    int
	Position       int
	ConnectedAt    time.Time
	DisconnectedAt *time.Time
	IsConnected    bool
}

type State struct {
	Session      string    `json:"session"`
	SessionStart time.Time `json:"sessionStart"`
	PlayerCount  int       `json:"playerCount"`
}

type ServerState struct {
	sync.RWMutex           `swaggerignore:"-" json:"-"`
	Session                TrackSession `json:"session"`
	SessionStart           time.Time    `json:"sessionStart"`
	PlayerCount            int          `json:"playerCount"`
	Track                  string       `json:"track"`
	MaxConnections         int          `json:"maxConnections"`
	SessionDurationMinutes int          `json:"sessionDurationMinutes"`
}

type ServerFilter struct {
	BaseFilter
	ServerBasedFilter
	Name        string `query:"name"`
	ServiceName string `query:"service_name"`
	Status      string `query:"status"`
}

func (f *ServerFilter) ApplyFilter(query *gorm.DB) *gorm.DB {
	if f.ServerID != "" {
		if serverUUID, err := uuid.Parse(f.ServerID); err == nil {
			query = query.Where("id = ?", serverUUID)
		}
	}

	return query
}

func (s *Server) GenerateUUID() {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
}

func (s *Server) BeforeCreate(tx *gorm.DB) error {
	if s.Name == "" {
		return errors.New("server name is required")
	}

	s.GenerateUUID()

	if s.ServiceName == "" {
		s.ServiceName = s.GenerateServiceName()
	}
	if s.Path == "" {
		s.Path = s.GenerateServerPath(BaseServerPath)
	}

	if s.DateCreated.IsZero() {
		s.DateCreated = time.Now().UTC()
	}

	return nil
}

func (s *Server) GenerateServiceName() string {
	if s.ID != uuid.Nil {
		return fmt.Sprintf("%s-%s", ServiceNamePrefix, s.ID.String()[:8])
	}
	return fmt.Sprintf("%s-%d", ServiceNamePrefix, time.Now().UnixNano())
}

func (s *Server) GenerateServerPath(steamCMDPath string) string {
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
