package model

import (
	"acc-server-manager/local/utl/password"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user account in the system.
type User struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Username string    `json:"username" gorm:"unique_index;not null"`
	Password string    `json:"-" gorm:"not null"` // Never expose password in JSON
	RoleID   uuid.UUID `json:"role_id" gorm:"type:uuid"`
	Role     Role      `json:"role"`
}

// BeforeCreate is a GORM hook that runs before creating new users
func (s *User) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()

	// Validate password strength
	if err := password.ValidatePasswordStrength(s.Password); err != nil {
		return err
	}

	// Hash password before saving
	hashed, err := password.HashPassword(s.Password)
	if err != nil {
		return err
	}
	s.Password = hashed

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating users
func (s *User) BeforeUpdate(tx *gorm.DB) error {
	// Only hash if password field is being updated
	if tx.Statement.Changed("Password") {
		// Validate password strength
		if err := password.ValidatePasswordStrength(s.Password); err != nil {
			return err
		}

		hashed, err := password.HashPassword(s.Password)
		if err != nil {
			return err
		}
		s.Password = hashed
	}

	return nil
}

// AfterFind is a GORM hook that runs after fetching users
func (s *User) AfterFind(tx *gorm.DB) error {
	// Password remains hashed - never decrypt
	// This hook is kept for potential future use
	return nil
}

// Validate checks if the user data is valid
func (s *User) Validate() error {
	if s.Username == "" {
		return errors.New("username is required")
	}
	if s.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// VerifyPassword verifies a plain text password against the stored hash
func (s *User) VerifyPassword(plainPassword string) error {
	return password.VerifyPassword(s.Password, plainPassword)
}
