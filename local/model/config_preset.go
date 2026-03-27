package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConfigPreset struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Name          string    `json:"name" gorm:"not null"`
	Description   string    `json:"description"`
	Configuration string    `json:"configuration" gorm:"type:text"`
	AssistRules   string    `json:"assistRules" gorm:"type:text"`
	Event         string    `json:"event" gorm:"type:text"`
	EventRules    string    `json:"eventRules" gorm:"type:text"`
	Settings      string    `json:"settings" gorm:"type:text"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (p *ConfigPreset) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type ConfigPresetCreateRequest struct {
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	Configuration *map[string]interface{} `json:"configuration"`
	AssistRules   *map[string]interface{} `json:"assistRules"`
	Event         *map[string]interface{} `json:"event"`
	EventRules    *map[string]interface{} `json:"eventRules"`
	Settings      *map[string]interface{} `json:"settings"`
}

type ConfigPresetUpdateRequest struct {
	Name          *string                 `json:"name"`
	Description   *string                 `json:"description"`
	Configuration *map[string]interface{} `json:"configuration"`
	AssistRules   *map[string]interface{} `json:"assistRules"`
	Event         *map[string]interface{} `json:"event"`
	EventRules    *map[string]interface{} `json:"eventRules"`
	Settings      *map[string]interface{} `json:"settings"`
}
