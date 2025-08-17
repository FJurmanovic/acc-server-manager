package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/jwt"
	"acc-server-manager/local/utl/password"
	"acc-server-manager/tests"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestJWT_GenerateAndValidateToken(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	jwtHandler := jwt.NewJWTHandler(os.Getenv("JWT_SECRET"))

	// Create test user
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		RoleID:   uuid.New(),
	}

	// Test JWT generation
	token, err := jwtHandler.GenerateToken(user.ID.String())
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, token)

	// Verify token is not empty
	if token == "" {
		t.Fatal("Expected non-empty token, got empty string")
	}

	// Test JWT validation
	claims, err := jwtHandler.ValidateToken(token)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, claims)
	tests.AssertEqual(t, user.ID.String(), claims.UserID)
}

func TestJWT_ValidateToken_InvalidToken(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()
	jwtHandler := jwt.NewJWTHandler(os.Getenv("JWT_SECRET"))

	// Test with invalid token
	claims, err := jwtHandler.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
	// Direct nil check to avoid the interface wrapping issue
	if claims != nil {
		t.Fatalf("Expected nil claims, got %v", claims)
	}
}

func TestJWT_ValidateToken_EmptyToken(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()
	jwtHandler := jwt.NewJWTHandler(os.Getenv("JWT_SECRET"))

	// Test with empty token
	claims, err := jwtHandler.ValidateToken("")
	if err == nil {
		t.Fatal("Expected error for empty token, got nil")
	}
	// Direct nil check to avoid the interface wrapping issue
	if claims != nil {
		t.Fatalf("Expected nil claims, got %v", claims)
	}
}

func TestUser_VerifyPassword_Success(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create test user
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		RoleID:   uuid.New(),
	}

	// Hash password manually (simulating what BeforeCreate would do)
	plainPassword := "password123"
	hashedPassword, err := password.HashPassword(plainPassword)
	tests.AssertNoError(t, err)
	user.Password = hashedPassword

	// Test password verification - should succeed
	err = user.VerifyPassword(plainPassword)
	tests.AssertNoError(t, err)
}

func TestUser_VerifyPassword_Failure(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create test user
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		RoleID:   uuid.New(),
	}

	// Hash password manually
	hashedPassword, err := password.HashPassword("correct_password")
	tests.AssertNoError(t, err)
	user.Password = hashedPassword

	// Test password verification with wrong password - should fail
	err = user.VerifyPassword("wrong_password")
	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}
}

func TestUser_Validate_Success(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create valid user
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: "password123",
		RoleID:   uuid.New(),
	}

	// Test validation - should succeed
	err := user.Validate()
	tests.AssertNoError(t, err)
}

func TestUser_Validate_MissingUsername(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create user without username
	user := &model.User{
		ID:       uuid.New(),
		Username: "", // Missing username
		Password: "password123",
		RoleID:   uuid.New(),
	}

	// Test validation - should fail
	err := user.Validate()
	if err == nil {
		t.Fatal("Expected error for missing username, got nil")
	}
}

func TestUser_Validate_MissingPassword(t *testing.T) {
	// Setup
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	// Create user without password
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: "", // Missing password
		RoleID:   uuid.New(),
	}

	// Test validation - should fail
	err := user.Validate()
	if err == nil {
		t.Fatal("Expected error for missing password, got nil")
	}
}

func TestPassword_HashAndVerify(t *testing.T) {
	// Test password hashing and verification directly
	plainPassword := "test_password_123"

	// Hash password
	hashedPassword, err := password.HashPassword(plainPassword)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, hashedPassword)

	// Verify hashed password is not the same as plain password
	if hashedPassword == plainPassword {
		t.Fatal("Hashed password should not equal plain password")
	}

	// Verify correct password
	err = password.VerifyPassword(hashedPassword, plainPassword)
	tests.AssertNoError(t, err)

	// Verify wrong password fails
	err = password.VerifyPassword(hashedPassword, "wrong_password")
	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}
}

func TestPassword_ValidatePasswordStrength(t *testing.T) {
	testCases := []struct {
		name        string
		password    string
		shouldError bool
	}{
		{"Valid password", "StrongPassword123!", false},
		{"Too short", "123", true},
		{"Empty password", "", true},
		{"Medium password", "password123", false}, // Depends on validation rules
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := password.ValidatePasswordStrength(tc.password)
			if tc.shouldError {
				if err == nil {
					t.Fatalf("Expected error for password '%s', got nil", tc.password)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error for password '%s', got: %v", tc.password, err)
				}
			}
		})
	}
}

func TestRole_Model(t *testing.T) {
	// Test Role model structure
	permissions := []model.Permission{
		{ID: uuid.New(), Name: "read"},
		{ID: uuid.New(), Name: "write"},
		{ID: uuid.New(), Name: "admin"},
	}

	role := &model.Role{
		ID:          uuid.New(),
		Name:        "Test Role",
		Permissions: permissions,
	}

	// Verify role structure
	tests.AssertEqual(t, "Test Role", role.Name)
	tests.AssertEqual(t, 3, len(role.Permissions))
	tests.AssertEqual(t, "read", role.Permissions[0].Name)
	tests.AssertEqual(t, "write", role.Permissions[1].Name)
	tests.AssertEqual(t, "admin", role.Permissions[2].Name)
}

func TestPermission_Model(t *testing.T) {
	// Test Permission model structure
	permission := &model.Permission{
		ID:   uuid.New(),
		Name: "test_permission",
	}

	// Verify permission structure
	tests.AssertEqual(t, "test_permission", permission.Name)
	tests.AssertNotNil(t, permission.ID)
}

func TestUser_WithRole_Model(t *testing.T) {
	// Test User model with Role relationship
	permissions := []model.Permission{
		{ID: uuid.New(), Name: "read"},
		{ID: uuid.New(), Name: "write"},
	}

	role := model.Role{
		ID:          uuid.New(),
		Name:        "User",
		Permissions: permissions,
	}

	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: "hashedpassword",
		RoleID:   role.ID,
		Role:     role,
	}

	// Verify user-role relationship
	tests.AssertEqual(t, "testuser", user.Username)
	tests.AssertEqual(t, role.ID, user.RoleID)
	tests.AssertEqual(t, "User", user.Role.Name)
	tests.AssertEqual(t, 2, len(user.Role.Permissions))
	tests.AssertEqual(t, "read", user.Role.Permissions[0].Name)
	tests.AssertEqual(t, "write", user.Role.Permissions[1].Name)
}
