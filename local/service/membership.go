package service

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/repository"
	"acc-server-manager/local/utl/jwt"
	"acc-server-manager/local/utl/logging"
	"context"
	"errors"
	"os"

	"github.com/google/uuid"
)

// MembershipService provides business logic for membership-related operations.
type MembershipService struct {
	repo *repository.MembershipRepository
}

// NewMembershipService creates a new MembershipService.
func NewMembershipService(repo *repository.MembershipRepository) *MembershipService {
	return &MembershipService{
		repo: repo,
	}
}

// Login authenticates a user and returns a JWT.
func (s *MembershipService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Use secure password verification with constant-time comparison
	if err := user.VerifyPassword(password); err != nil {
		return "", errors.New("invalid credentials")
	}

	return jwt.GenerateToken(user)
}

// CreateUser creates a new user.
func (s *MembershipService) CreateUser(ctx context.Context, username, password, roleName string) (*model.User, error) {

	role, err := s.repo.FindRoleByName(ctx, roleName)
	if err != nil {
		logging.Error("Failed to find role by name: %v", err)
		return nil, errors.New("role not found")
	}

	user := &model.User{
		Username: username,
		Password: password,
		RoleID:   role.ID,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		logging.Error("Failed to create user: %v", err)
		return nil, err
	}

	logging.InfoOperation("USER_CREATE", "Created user: "+user.Username+" (ID: "+user.ID.String()+", Role: "+roleName+")")
	return user, nil
}

// ListUsers retrieves all users.
func (s *MembershipService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.repo.ListUsers(ctx)
}

// GetUser retrieves a single user by ID.
func (s *MembershipService) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.repo.FindUserByID(ctx, userID)
}

// GetUserWithPermissions retrieves a single user by ID with their role and permissions.
func (s *MembershipService) GetUserWithPermissions(ctx context.Context, userID string) (*model.User, error) {
	return s.repo.FindUserByIDWithPermissions(ctx, userID)
}

// UpdateUserRequest defines the request body for updating a user.
type UpdateUserRequest struct {
	Username *string    `json:"username"`
	Password *string    `json:"password"`
	RoleID   *uuid.UUID `json:"roleId"`
}

// DeleteUser deletes a user with validation to prevent Super Admin deletion.
func (s *MembershipService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	// Get user with role information
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Get role to check if it's Super Admin
	role, err := s.repo.FindRoleByID(ctx, user.RoleID)
	if err != nil {
		return errors.New("user role not found")
	}

	// Prevent deletion of Super Admin users
	if role.Name == "Super Admin" {
		return errors.New("cannot delete Super Admin user")
	}

	err = s.repo.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	logging.InfoOperation("USER_DELETE", "Deleted user: "+userID.String())
	return nil
}

// UpdateUser updates a user's details.
func (s *MembershipService) UpdateUser(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Username != nil {
		user.Username = *req.Username
	}

	if req.Password != nil && *req.Password != "" {
		// Password will be automatically hashed in BeforeUpdate hook
		user.Password = *req.Password
	}

	if req.RoleID != nil {
		// Check if role exists
		_, err := s.repo.FindRoleByID(ctx, *req.RoleID)
		if err != nil {
			return nil, errors.New("role not found")
		}
		user.RoleID = *req.RoleID
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	logging.InfoOperation("USER_UPDATE", "Updated user: "+user.Username+" (ID: "+user.ID.String()+")")
	return user, nil
}

// HasPermission checks if a user has a specific permission.
func (s *MembershipService) HasPermission(ctx context.Context, userID string, permissionName string) (bool, error) {
	user, err := s.repo.FindUserByIDWithPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Super admin and Admin have all permissions
	if user.Role.Name == "Super Admin" || user.Role.Name == "Admin" {
		return true, nil
	}

	for _, p := range user.Role.Permissions {
		if p.Name == permissionName {
			return true, nil
		}
	}

	return false, nil
}

// SetupInitialData creates the initial roles and permissions.
func (s *MembershipService) SetupInitialData(ctx context.Context) error {
	// Define all permissions
	permissions := model.AllPermissions()

	createdPermissions := make([]model.Permission, 0)
	for _, pName := range permissions {
		perm, err := s.repo.FindPermissionByName(ctx, pName)
		if err != nil { // Assuming error means not found
			perm = &model.Permission{Name: pName}
			if err := s.repo.CreatePermission(ctx, perm); err != nil {
				return err
			}
		}
		createdPermissions = append(createdPermissions, *perm)
	}

	// Create Super Admin role with all permissions
	superAdminRole, err := s.repo.FindRoleByName(ctx, "Super Admin")
	if err != nil {
		superAdminRole = &model.Role{Name: "Super Admin"}
		if err := s.repo.CreateRole(ctx, superAdminRole); err != nil {
			return err
		}
	}
	if err := s.repo.AssignPermissionsToRole(ctx, superAdminRole, createdPermissions); err != nil {
		return err
	}

	// Create Admin role with same permissions as Super Admin
	adminRole, err := s.repo.FindRoleByName(ctx, "Admin")
	if err != nil {
		adminRole = &model.Role{Name: "Admin"}
		if err := s.repo.CreateRole(ctx, adminRole); err != nil {
			return err
		}
	}
	if err := s.repo.AssignPermissionsToRole(ctx, adminRole, createdPermissions); err != nil {
		return err
	}

	// Create Manager role with limited permissions (excluding membership, role, user, server create/delete)
	managerRole, err := s.repo.FindRoleByName(ctx, "Manager")
	if err != nil {
		managerRole = &model.Role{Name: "Manager"}
		if err := s.repo.CreateRole(ctx, managerRole); err != nil {
			return err
		}
	}

	// Define manager permissions (limited set)
	managerPermissionNames := []string{
		model.ServerView,
		model.ServerUpdate,
		model.ServerStart,
		model.ServerStop,
		model.ConfigView,
		model.ConfigUpdate,
	}

	managerPermissions := make([]model.Permission, 0)
	for _, permName := range managerPermissionNames {
		for _, perm := range createdPermissions {
			if perm.Name == permName {
				managerPermissions = append(managerPermissions, perm)
				break
			}
		}
	}

	if err := s.repo.AssignPermissionsToRole(ctx, managerRole, managerPermissions); err != nil {
		return err
	}

	// Create a default admin user if one doesn't exist
	_, err = s.repo.FindUserByUsername(ctx, "admin")
	if err != nil {
		logging.Debug("Creating default admin user")
		_, err = s.CreateUser(ctx, "admin", os.Getenv("PASSWORD"), "Super Admin") // Default password, should be changed
		if err != nil {
			return err
		}
	}

	return nil
}

// GetAllRoles retrieves all roles for dropdown selection.
func (s *MembershipService) GetAllRoles(ctx context.Context) ([]*model.Role, error) {
	return s.repo.ListRoles(ctx)
}
