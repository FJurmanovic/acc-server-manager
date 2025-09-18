package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;"`
	Name        string       `json:"name" gorm:"unique_index;not null"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

func (s *Role) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()

	return nil
}
