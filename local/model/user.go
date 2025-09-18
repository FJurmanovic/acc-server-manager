package model

import (
	"acc-server-manager/local/utl/password"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;"`
	Username string    `json:"username" gorm:"unique_index;not null"`
	Password string    `json:"-" gorm:"not null"` // Never expose password in JSON
	RoleID   uuid.UUID `json:"role_id" gorm:"type:uuid"`
	Role     Role      `json:"role"`
}

func (s *User) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New()

	if err := password.ValidatePasswordStrength(s.Password); err != nil {
		return err
	}

	hashed, err := password.HashPassword(s.Password)
	if err != nil {
		return err
	}
	s.Password = hashed

	return nil
}

func (s *User) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("Password") {
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

func (s *User) AfterFind(tx *gorm.DB) error {
	return nil
}

func (s *User) Validate() error {
	if s.Username == "" {
		return errors.New("username is required")
	}
	if s.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func (s *User) VerifyPassword(plainPassword string) error {
	return password.VerifyPassword(s.Password, plainPassword)
}
