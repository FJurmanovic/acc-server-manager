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
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	jwtHandler := jwt.NewJWTHandler(os.Getenv("JWT_SECRET"))
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		RoleID:   uuid.New(),
	}

	token, err := jwtHandler.GenerateToken(user.ID.String())
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, token)

	if token == "" {
		t.Fatal("Expected non-empty token, got empty string")
	}

	claims, err := jwtHandler.ValidateToken(token)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, claims)
	tests.AssertEqual(t, user.ID.String(), claims.UserID)
}

func TestJWT_ValidateToken_InvalidToken(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()
	jwtHandler := jwt.NewJWTHandler(os.Getenv("JWT_SECRET"))

	claims, err := jwtHandler.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
	if claims != nil {
		t.Fatalf("Expected nil claims, got %v", claims)
	}
}

func TestJWT_ValidateToken_EmptyToken(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()
	jwtHandler := jwt.NewJWTHandler(os.Getenv("JWT_SECRET"))

	claims, err := jwtHandler.ValidateToken("")
	if err == nil {
		t.Fatal("Expected error for empty token, got nil")
	}
	if claims != nil {
		t.Fatalf("Expected nil claims, got %v", claims)
	}
}

func TestUser_VerifyPassword_Success(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		RoleID:   uuid.New(),
	}

	plainPassword := "password123"
	hashedPassword, err := password.HashPassword(plainPassword)
	tests.AssertNoError(t, err)
	user.Password = hashedPassword

	err = user.VerifyPassword(plainPassword)
	tests.AssertNoError(t, err)
}

func TestUser_VerifyPassword_Failure(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		RoleID:   uuid.New(),
	}

	hashedPassword, err := password.HashPassword("correct_password")
	tests.AssertNoError(t, err)
	user.Password = hashedPassword

	err = user.VerifyPassword("wrong_password")
	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}
}

func TestUser_Validate_Success(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: "password123",
		RoleID:   uuid.New(),
	}

	err := user.Validate()
	tests.AssertNoError(t, err)
}

func TestUser_Validate_MissingUsername(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	user := &model.User{
		ID:       uuid.New(),
		Username: "",
		Password: "password123",
		RoleID:   uuid.New(),
	}

	err := user.Validate()
	if err == nil {
		t.Fatal("Expected error for missing username, got nil")
	}
}

func TestUser_Validate_MissingPassword(t *testing.T) {
	helper := tests.NewTestHelper(t)
	defer helper.Cleanup()

	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		Password: "",
		RoleID:   uuid.New(),
	}

	err := user.Validate()
	if err == nil {
		t.Fatal("Expected error for missing password, got nil")
	}
}

func TestPassword_HashAndVerify(t *testing.T) {
	plainPassword := "test_password_123"

	hashedPassword, err := password.HashPassword(plainPassword)
	tests.AssertNoError(t, err)
	tests.AssertNotNil(t, hashedPassword)

	if hashedPassword == plainPassword {
		t.Fatal("Hashed password should not equal plain password")
	}

	err = password.VerifyPassword(hashedPassword, plainPassword)
	tests.AssertNoError(t, err)

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
		{"Medium password", "password123", false},
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

	tests.AssertEqual(t, "Test Role", role.Name)
	tests.AssertEqual(t, 3, len(role.Permissions))
	tests.AssertEqual(t, "read", role.Permissions[0].Name)
	tests.AssertEqual(t, "write", role.Permissions[1].Name)
	tests.AssertEqual(t, "admin", role.Permissions[2].Name)
}

func TestPermission_Model(t *testing.T) {
	permission := &model.Permission{
		ID:   uuid.New(),
		Name: "test_permission",
	}

	tests.AssertEqual(t, "test_permission", permission.Name)
	tests.AssertNotNil(t, permission.ID)
}

func TestUser_WithRole_Model(t *testing.T) {
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

	tests.AssertEqual(t, "testuser", user.Username)
	tests.AssertEqual(t, role.ID, user.RoleID)
	tests.AssertEqual(t, "User", user.Role.Name)
	tests.AssertEqual(t, 2, len(user.Role.Permissions))
	tests.AssertEqual(t, "read", user.Role.Permissions[0].Name)
	tests.AssertEqual(t, "write", user.Role.Permissions[1].Name)
}
