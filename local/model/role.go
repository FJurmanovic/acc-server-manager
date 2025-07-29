package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents a user role in the system.
type Role struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;"`
	Name        string       `json:"name" gorm:"unique_index;not null"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

// BeforeCreate is a GORM hook that runs before creating new credentials
func (s *Role) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()

	return nil
}