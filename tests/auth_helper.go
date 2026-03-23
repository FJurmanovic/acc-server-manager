package tests

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/jwt"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

func GenerateTestToken() (string, error) {
	user := &model.User{
		ID:       uuid.New(),
		Username: "test_user",
		RoleID:   uuid.New(),
	}

	testSecret := os.Getenv("JWT_SECRET")
	if testSecret == "" {
		testSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(testSecret)

	token, err := jwtHandler.GenerateToken(user.ID.String())
	if err != nil {
		return "", fmt.Errorf("failed to generate test token: %w", err)
	}

	return token, nil
}

func MustGenerateTestToken() string {
	token, err := GenerateTestToken()
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test token: %v", err))
	}
	return token
}

func GenerateTestTokenWithExpiry(expiryTime time.Time) (string, error) {
	testSecret := os.Getenv("JWT_SECRET")
	if testSecret == "" {
		testSecret = "test-secret-that-is-at-least-32-bytes-long-for-security"
	}
	jwtHandler := jwt.NewJWTHandler(testSecret)

	user := &model.User{
		ID:       uuid.New(),
		Username: "test_user",
		RoleID:   uuid.New(),
	}

	token, err := jwtHandler.GenerateTokenWithExpiry(user, expiryTime)
	if err != nil {
		return "", fmt.Errorf("failed to generate test token with expiry: %w", err)
	}

	return token, nil
}

func AddAuthHeader(headers map[string]string) (map[string]string, error) {
	token, err := GenerateTestToken()
	if err != nil {
		return nil, err
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Authorization"] = "Bearer " + token

	return headers, nil
}

func MustAddAuthHeader(headers map[string]string) map[string]string {
	result, err := AddAuthHeader(headers)
	if err != nil {
		panic(fmt.Sprintf("Failed to add auth header: %v", err))
	}
	return result
}
