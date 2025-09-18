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

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SteamCredentials struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	Username    string    `gorm:"not null" json:"username"`
	Password    string    `gorm:"not null" json:"-"`
	DateCreated time.Time `json:"dateCreated"`
	LastUpdated time.Time `json:"lastUpdated"`
}

func (SteamCredentials) TableName() string {
	return "steam_credentials"
}

func (s *SteamCredentials) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	now := time.Now().UTC()
	if s.DateCreated.IsZero() {
		s.DateCreated = now
	}
	s.LastUpdated = now

	encrypted, err := EncryptPassword(s.Password)
	if err != nil {
		return err
	}
	s.Password = encrypted

	return nil
}

func (s *SteamCredentials) BeforeUpdate(tx *gorm.DB) error {
	s.LastUpdated = time.Now().UTC()

	if tx.Statement.Changed("Password") {
		encrypted, err := EncryptPassword(s.Password)
		if err != nil {
			return err
		}
		s.Password = encrypted
	}

	return nil
}

func (s *SteamCredentials) AfterFind(tx *gorm.DB) error {
	if s.Password != "" {
		decrypted, err := DecryptPassword(s.Password)
		if err != nil {
			return err
		}
		s.Password = decrypted
	}
	return nil
}

func (s *SteamCredentials) Validate() error {
	if s.Username == "" {
		return errors.New("username is required")
	}

	if len(s.Username) < 3 || len(s.Username) > 64 {
		return errors.New("username must be between 3 and 64 characters")
	}

	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, s.Username); !matched {
		return errors.New("username contains invalid characters")
	}

	if s.Password == "" {
		return errors.New("password is required")
	}

	if len(s.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	if len(s.Password) > 128 {
		return errors.New("password is too long")
	}

	weakPasswords := []string{"password", "123456", "steam", "admin", "user"}
	lowerPass := strings.ToLower(s.Password)
	for _, weak := range weakPasswords {
		if lowerPass == weak {
			return errors.New("password is too weak")
		}
	}

	return nil
}

func GetEncryptionKey() []byte {
	key := []byte(configs.EncryptionKey)
	if len(key) != 32 {
		panic("encryption key must be exactly 32 bytes for AES-256")
	}
	return key
}

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

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", errors.New("encrypted password cannot be empty")
	}

	if len(encryptedPassword) < 24 {
		return "", errors.New("invalid encrypted password format")
	}

	key := GetEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

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

	decrypted := string(plaintext)
	if len(decrypted) == 0 || len(decrypted) > 1024 {
		return "", errors.New("invalid decrypted password")
	}

	return decrypted, nil
}
