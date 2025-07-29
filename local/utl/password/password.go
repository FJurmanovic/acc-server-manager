package password

import (
	"errors"
	"os"

	"golang.org/x/crypto/bcrypt"
)

const (
	// MinPasswordLength defines the minimum password length
	MinPasswordLength = 8
	// BcryptCost defines the cost factor for bcrypt hashing
	BcryptCost = 12
)

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", errors.New("password must be at least 8 characters long")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a plain text password against a hashed password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePasswordStrength validates password complexity requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return errors.New("password must be at least 8 characters long")
	}

	if os.Getenv("ENFORCE_PASSWORD_STRENGTH") == "true" {
		if len(password) < MinPasswordLength {
			return errors.New("password must be at least 8 characters long")
		}

		hasUpper := false
		hasLower := false
		hasDigit := false
		hasSpecial := false

		for _, char := range password {
			switch {
			case char >= 'A' && char <= 'Z':
				hasUpper = true
			case char >= 'a' && char <= 'z':
				hasLower = true
			case char >= '0' && char <= '9':
				hasDigit = true
			case char >= '!' && char <= '/' || char >= ':' && char <= '@' || char >= '[' && char <= '`' || char >= '{' && char <= '~':
				hasSpecial = true
			}
		}

		if !hasUpper {
			return errors.New("password must contain at least one uppercase letter")
		}
		if !hasLower {
			return errors.New("password must contain at least one lowercase letter")
		}
		if !hasDigit {
			return errors.New("password must contain at least one digit")
		}
		if !hasSpecial {
			return errors.New("password must contain at least one special character")
		}

		return nil
	}

	return nil
}
