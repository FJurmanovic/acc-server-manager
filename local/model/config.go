package model

import "time"

// Config tracks configuration modifications
type Config struct {
	ID         uint      `gorm:"primaryKey"`
	ServerID   uint      `gorm:"not null"`
	ConfigFile string    `gorm:"not null"` // e.g. "settings.json"
	OldConfig  string    `gorm:"type:text"`
	NewConfig  string    `gorm:"type:text"`
	ChangedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type Configurations struct {
	Configuration map[string]interface{}
	Entrylist     map[string]interface{}
	Event         map[string]interface{}
	EventRules    map[string]interface{}
	Settings      map[string]interface{}
}
