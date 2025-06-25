package jwt

import (
	"acc-server-manager/local/model"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// SecretKey is the secret key for signing the JWT.
// It is recommended to use a long, complex string for this.
// In a production environment, this should be loaded from a secure configuration source.
var SecretKey = []byte("your-secret-key")

// Claims represents the JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
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
