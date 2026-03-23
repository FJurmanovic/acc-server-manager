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

type CacheInvalidator interface {
	InvalidateUserPermissions(userID string)
	InvalidateAllUserPermissions()
}

type MembershipService struct {
	repo             *repository.MembershipRepository
	cacheInvalidator CacheInvalidator
	jwtHandler       *jwt.JWTHandler
	openJwtHandler   *jwt.OpenJWTHandler
}

func NewMembershipService(repo *repository.MembershipRepository, jwtHandler *jwt.JWTHandler, openJwtHandler *jwt.OpenJWTHandler) *MembershipService {
	return &MembershipService{
		repo:             repo,
		cacheInvalidator: nil,
		jwtHandler:       jwtHandler,
		openJwtHandler:   openJwtHandler,
	}
}

func (s *MembershipService) SetCacheInvalidator(invalidator CacheInvalidator) {
	s.cacheInvalidator = invalidator
}

func (s *MembershipService) HandleLogin(ctx context.Context, username, password string) (*model.User, error) {
	user, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := user.VerifyPassword(password); err != nil {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

func (s *MembershipService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.HandleLogin(ctx, username, password)
	if err != nil {
		return "", err
	}

	return s.jwtHandler.GenerateToken(user.ID.String())
}

func (s *MembershipService) GenerateOpenToken(ctx context.Context, userId string) (string, error) {
	return s.openJwtHandler.GenerateToken(userId)
}

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

func (s *MembershipService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.repo.ListUsers(ctx)
}

func (s *MembershipService) GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	return s.repo.FindUserByID(ctx, userID)
}

func (s *MembershipService) GetUserWithPermissions(ctx context.Context, userID string) (*model.User, error) {
	return s.repo.FindUserByIDWithPermissions(ctx, userID)
}

type UpdateUserRequest struct {
	Username *string    `json:"username"`
	Password *string    `json:"password"`
	RoleID   *uuid.UUID `json:"roleId"`
}

func (s *MembershipService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	role, err := s.repo.FindRoleByID(ctx, user.RoleID)
	if err != nil {
		return errors.New("user role not found")
	}

	if role.Name == "Super Admin" {
		return errors.New("cannot delete Super Admin user")
	}

	err = s.repo.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	if s.cacheInvalidator != nil {
		s.cacheInvalidator.InvalidateUserPermissions(userID.String())
	}

	logging.InfoOperation("USER_DELETE", "Deleted user: "+userID.String())
	return nil
}

func (s *MembershipService) UpdateUser(ctx context.Context, userID uuid.UUID, req UpdateUserRequest) (*model.User, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Username != nil {
		user.Username = *req.Username
	}

	if req.Password != nil && *req.Password != "" {
		user.Password = *req.Password
	}

	if req.RoleID != nil {
		_, err := s.repo.FindRoleByID(ctx, *req.RoleID)
		if err != nil {
			return nil, errors.New("role not found")
		}
		user.RoleID = *req.RoleID
	}

	if err := s.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	if req.RoleID != nil && s.cacheInvalidator != nil {
		s.cacheInvalidator.InvalidateUserPermissions(userID.String())
	}

	logging.InfoOperation("USER_UPDATE", "Updated user: "+user.Username+" (ID: "+user.ID.String()+")")
	return user, nil
}

func (s *MembershipService) HasPermission(ctx context.Context, userID string, permissionName string) (bool, error) {
	user, err := s.repo.FindUserByIDWithPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

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

func (s *MembershipService) SetupInitialData(ctx context.Context) error {
	permissions := model.AllPermissions()

	createdPermissions := make([]model.Permission, 0)
	for _, pName := range permissions {
		perm, err := s.repo.FindPermissionByName(ctx, pName)
		if err != nil {
			perm = &model.Permission{Name: pName}
			if err := s.repo.CreatePermission(ctx, perm); err != nil {
				return err
			}
		}
		createdPermissions = append(createdPermissions, *perm)
	}

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

	managerRole, err := s.repo.FindRoleByName(ctx, "Manager")
	if err != nil {
		managerRole = &model.Role{Name: "Manager"}
		if err := s.repo.CreateRole(ctx, managerRole); err != nil {
			return err
		}
	}

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

	if s.cacheInvalidator != nil {
		s.cacheInvalidator.InvalidateAllUserPermissions()
	}

	_, err = s.repo.FindUserByUsername(ctx, "admin")
	if err != nil {
		logging.Debug("Creating default admin user")
		_, err = s.CreateUser(ctx, "admin", os.Getenv("PASSWORD"), "Super Admin")
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *MembershipService) GetAllRoles(ctx context.Context) ([]*model.Role, error) {
	return s.repo.ListRoles(ctx)
}
