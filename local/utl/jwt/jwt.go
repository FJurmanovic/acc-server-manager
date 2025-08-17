package jwt

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/errors"
	"crypto/rand"
	"encoding/base64"
	goerrors "errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims represents the JWT claims.
type Claims struct {
	UserID      string `json:"user_id"`
	IsOpenToken bool   `json:"is_open_token"`
	jwt.RegisteredClaims
}

type JWTHandler struct {
	SecretKey   []byte
	IsOpenToken bool
}

type OpenJWTHandler struct {
	*JWTHandler
}

// NewJWTHandler creates a new JWTHandler instance with the provided secret key.
func NewOpenJWTHandler(jwtSecret string) *OpenJWTHandler {
	jwtHandler := NewJWTHandler(jwtSecret)
	jwtHandler.IsOpenToken = true
	return &OpenJWTHandler{
		JWTHandler: jwtHandler,
	}
}

// NewJWTHandler creates a new JWTHandler instance with the provided secret key.
func NewJWTHandler(jwtSecret string) *JWTHandler {
	if jwtSecret == "" {
		errors.SafeFatal("JWT_SECRET environment variable is required and cannot be empty")
	}

	var secretKey []byte

	// Decode base64 secret if it looks like base64, otherwise use as-is
	if decoded, err := base64.StdEncoding.DecodeString(jwtSecret); err == nil && len(decoded) >= 32 {
		secretKey = decoded
	} else {
		secretKey = []byte(jwtSecret)
	}

	// Ensure minimum key length for security
	if len(secretKey) < 32 {
		errors.SafeFatal("JWT_SECRET must be at least 32 bytes long for security")
	}
	return &JWTHandler{
		SecretKey: secretKey,
	}
}

// GenerateSecretKey generates a cryptographically secure random key for JWT signing
// This is a utility function for generating new secrets, not used in normal operation
func (jh *JWTHandler) GenerateSecretKey() string {
	key := make([]byte, 64) // 512 bits
	if _, err := rand.Read(key); err != nil {
		errors.SafeFatal("Failed to generate random key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

// GenerateToken generates a new JWT for a given user.
func (jh *JWTHandler) GenerateToken(userId string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		IsOpenToken: jh.IsOpenToken,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jh.SecretKey)
}

func (jh *JWTHandler) GenerateTokenWithExpiry(user *model.User, expiry time.Time) (string, error) {
	expirationTime := expiry
	claims := &Claims{
		UserID: user.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		IsOpenToken: jh.IsOpenToken,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jh.SecretKey)
}

// ValidateToken validates a JWT and returns the claims if the token is valid.
func (jh *JWTHandler) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jh.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, goerrors.New("invalid token")
	}

	return claims, nil
}
