package model

import (
	"acc-server-manager/local/utl/configs"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"gorm.io/gorm"
)

// SteamCredentials represents stored Steam login credentials
type SteamCredentials struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Username    string    `gorm:"not null" json:"username"`
	Password    string    `gorm:"not null" json:"-"` // Encrypted, not exposed in JSON
	DateCreated time.Time `json:"dateCreated"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// TableName specifies the table name for GORM
func (SteamCredentials) TableName() string {
	return "steam_credentials"
}

// BeforeCreate is a GORM hook that runs before creating new credentials
func (s *SteamCredentials) BeforeCreate(tx *gorm.DB) error {
	now := time.Now().UTC()
	if s.DateCreated.IsZero() {
		s.DateCreated = now
	}
	s.LastUpdated = now

	// Encrypt password before saving
	encrypted, err := EncryptPassword(s.Password)
	if err != nil {
		return err
	}
	s.Password = encrypted

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating credentials
func (s *SteamCredentials) BeforeUpdate(tx *gorm.DB) error {
	s.LastUpdated = time.Now().UTC()

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
func (s *SteamCredentials) AfterFind(tx *gorm.DB) error {
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
func (s *SteamCredentials) Validate() error {
	if s.Username == "" {
		return errors.New("username is required")
	}
	if s.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

// GetEncryptionKey returns the encryption key from config.
// The key is loaded from the ENCRYPTION_KEY environment variable.
func GetEncryptionKey() []byte {
	return []byte(configs.EncryptionKey)
}

// EncryptPassword encrypts a password using AES-256
func EncryptPassword(password string) (string, error) {
	key := GetEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the password
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)

	// Return base64 encoded encrypted password
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword decrypts an encrypted password
func DecryptPassword(encryptedPassword string) (string, error) {
	key := GetEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Decode base64 encoded password
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
} 