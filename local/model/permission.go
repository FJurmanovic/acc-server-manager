package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Permission represents an action that can be performed in the system.
type Permission struct {
	ID   uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Name string    `json:"name" gorm:"unique_index;not null"`
}

// BeforeCreate is a GORM hook that runs before creating new credentials
func (s *Permission) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()

	return nil
}