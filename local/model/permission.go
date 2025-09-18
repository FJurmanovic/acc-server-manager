package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Permission struct {
	ID   uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Name string    `json:"name" gorm:"unique_index;not null"`
}

func (s *Permission) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()

	return nil
}
