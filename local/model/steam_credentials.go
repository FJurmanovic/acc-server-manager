package model

import (
	"acc-server-manager/local/utl/configs"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"regexp"
	"strings"
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

// Validate checks if the credentials are valid with enhanced security checks
func (s *SteamCredentials) Validate() error {
	if s.Username == "" {
		return errors.New("username is required")
	}

	// Enhanced username validation
	if len(s.Username) < 3 || len(s.Username) > 64 {
		return errors.New("username must be between 3 and 64 characters")
	}

	// Check for valid characters in username (alphanumeric, underscore, hyphen)
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, s.Username); !matched {
		return errors.New("username contains invalid characters")
	}

	if s.Password == "" {
		return errors.New("password is required")
	}

	// Basic password validation
	if len(s.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	if len(s.Password) > 128 {
		return errors.New("password is too long")
	}

	// Check for obvious weak passwords
	weakPasswords := []string{"password", "123456", "steam", "admin", "user"}
	lowerPass := strings.ToLower(s.Password)
	for _, weak := range weakPasswords {
		if lowerPass == weak {
			return errors.New("password is too weak")
		}
	}

	return nil
}

// GetEncryptionKey returns the encryption key from config.
// The key is loaded from the ENCRYPTION_KEY environment variable.
func GetEncryptionKey() []byte {
	key := []byte(configs.EncryptionKey)
	if len(key) != 32 {
		panic("encryption key must be exactly 32 bytes for AES-256")
	}
	return key
}

// EncryptPassword encrypts a password using AES-256-GCM with enhanced security
func EncryptPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	if len(password) > 1024 {
		return "", errors.New("password too long")
	}

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

	// Create a cryptographically secure nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the password with authenticated encryption
	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)

	// Return base64 encoded encrypted password
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword decrypts an encrypted password with enhanced validation
func DecryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", errors.New("encrypted password cannot be empty")
	}

	// Validate base64 format
	if len(encryptedPassword) < 24 { // Minimum reasonable length
		return "", errors.New("invalid encrypted password format")
	}

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
		return "", errors.New("invalid base64 encoding")
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("decryption failed - invalid ciphertext or key")
	}

	// Validate decrypted content
	decrypted := string(plaintext)
	if len(decrypted) == 0 || len(decrypted) > 1024 {
		return "", errors.New("invalid decrypted password")
	}

	return decrypted, nil
}
