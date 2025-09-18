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

func NewOpenJWTHandler(jwtSecret string) *OpenJWTHandler {
	jwtHandler := NewJWTHandler(jwtSecret)
	jwtHandler.IsOpenToken = true
	return &OpenJWTHandler{
		JWTHandler: jwtHandler,
	}
}

func NewJWTHandler(jwtSecret string) *JWTHandler {
	if jwtSecret == "" {
		errors.SafeFatal("JWT_SECRET environment variable is required and cannot be empty")
	}

	var secretKey []byte

	if decoded, err := base64.StdEncoding.DecodeString(jwtSecret); err == nil && len(decoded) >= 32 {
		secretKey = decoded
	} else {
		secretKey = []byte(jwtSecret)
	}

	if len(secretKey) < 32 {
		errors.SafeFatal("JWT_SECRET must be at least 32 bytes long for security")
	}
	return &JWTHandler{
		SecretKey: secretKey,
	}
}

func (jh *JWTHandler) GenerateSecretKey() string {
	key := make([]byte, 64) // 512 bits
	if _, err := rand.Read(key); err != nil {
		errors.SafeFatal("Failed to generate random key: %v", err)
	}
	return base64.StdEncoding.EncodeToString(key)
}

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
