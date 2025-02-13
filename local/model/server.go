package model

// Server represents an ACC server instance
type Server struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Status      string `json:"status"`
	IP          string `gorm:"not null" json:"-"`
	Port        int    `gorm:"not null" json:"-"`
	ConfigPath  string `gorm:"not null" json:"-"` // e.g. "/acc/servers/server1/"
	ServiceName string `gorm:"not null" json:"-"` // Windows service name
}
