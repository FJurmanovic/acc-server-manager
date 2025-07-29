package tests

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/jwt"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GenerateTestToken creates a JWT token for testing purposes
func GenerateTestToken() (string, error) {
	// Create test user
	user := &model.User{
		ID:       uuid.New(),
		Username: "test_user",
		RoleID:   uuid.New(),
	}

	// Generate JWT token
	token, err := jwt.GenerateToken(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate test token: %w", err)
	}

	return token, nil
}

// MustGenerateTestToken generates a test token and panics if it fails
// This is useful for test setup where failing to generate a token is a fatal error
func MustGenerateTestToken() string {
	token, err := GenerateTestToken()
	if err != nil {
		panic(fmt.Sprintf("Failed to generate test token: %v", err))
	}
	return token
}

// GenerateTestTokenWithExpiry creates a JWT token with a specific expiry time
func GenerateTestTokenWithExpiry(expiryTime time.Time) (string, error) {
	// Create test user
	user := &model.User{
		ID:       uuid.New(),
		Username: "test_user",
		RoleID:   uuid.New(),
	}

	// Generate JWT token with custom expiry
	token, err := jwt.GenerateTokenWithExpiry(user, expiryTime)
	if err != nil {
		return "", fmt.Errorf("failed to generate test token with expiry: %w", err)
	}

	return token, nil
}

// AddAuthHeader adds a test auth token to the request headers
// This is a convenience method for tests that need to authenticate requests
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

// MustAddAuthHeader adds a test auth token to the request headers and panics if it fails
func MustAddAuthHeader(headers map[string]string) map[string]string {
	result, err := AddAuthHeader(headers)
	if err != nil {
		panic(fmt.Sprintf("Failed to add auth header: %v", err))
	}
	return result
}
