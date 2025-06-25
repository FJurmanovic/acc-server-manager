package model

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user account in the system.
type User struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Username string    `json:"username" gorm:"unique_index;not null"`
	Password string    `json:"password" gorm:"not null"`
	RoleID   uuid.UUID `json:"role_id" gorm:"type:uuid"`
	Role     Role      `json:"role"`
}


// BeforeCreate is a GORM hook that runs before creating new credentials
func (s *User) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()
	// Encrypt password before saving
	encrypted, err := EncryptPassword(s.Password)
	if err != nil {
		return err
	}
	s.Password = encrypted

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating credentials
func (s *User) BeforeUpdate(tx *gorm.DB) error {

	// Only encrypt if password field is being updated
	if tx.Statement.Changed("Password") {
		encrypted, err := EncryptPassword(s.Password)
		if err != nil {
			return err
		}
		s.Password = encrypted
	}

	return nil
}

// AfterFind is a GORM hook that runs after fetching credentials
func (s *User) AfterFind(tx *gorm.DB) error {
	// Decrypt password after fetching
	if s.Password != "" {
		decrypted, err := DecryptPassword(s.Password)
		if err != nil {
			return err
		}
		s.Password = decrypted
	}
	return nil
}

// Validate checks if the credentials are valid
func (s *User) Validate() error {
	if s.Username == "" {
		return errors.New("username is required")
	}
	if s.Password == "" {
		return errors.New("password is required")
	}
	return nil
}