package model

// Server represents an ACC server instance
type Server struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	IP          string `gorm:"not null"`
	Port        int    `gorm:"not null"`
	ConfigPath  string `gorm:"not null"` // e.g. "/acc/servers/server1/"
	ServiceName string `gorm:"not null"` // Windows service name
}
