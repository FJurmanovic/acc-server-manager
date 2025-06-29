package jwt

import (
	"acc-server-manager/local/model"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// SecretKey holds the JWT signing key loaded from environment
var SecretKey []byte

// Claims represents the JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// init initializes the JWT secret key from environment variable
func init() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required and cannot be empty")
	}

	// Decode base64 secret if it looks like base64, otherwise use as-is
	if decoded, err := base64.StdEncoding.DecodeString(jwtSecret); err == nil && len(decoded) >= 32 {
		SecretKey = decoded
	} else {
		SecretKey = []byte(jwtSecret)
	}

	// Ensure minimum key length for security
	if len(SecretKey) < 32 {
		log.Fatal("JWT_SECRET must be at least 32 bytes long for security")
	}
}

// GenerateSecretKey generates a cryptographically secure random key for JWT signing
// This is a utility function for generating new secrets, not used in normal operation
func GenerateSecretKey() string {
	key := make([]byte, 64) // 512 bits
	if _, err := rand.Read(key); err != nil {
		log.Fatal("Failed to generate random key: ", err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

// GenerateToken generates a new JWT for a given user.
func GenerateToken(user *model.User) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(SecretKey)
}

// ValidateToken validates a JWT and returns the claims if the token is valid.
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
