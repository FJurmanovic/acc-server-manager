package repository

import (
	"acc-server-manager/local/model"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MembershipRepository handles database operations for users, roles, and permissions.
type MembershipRepository struct {
	db *gorm.DB
}

// NewMembershipRepository creates a new MembershipRepository.
func NewMembershipRepository(db *gorm.DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

// FindUserByUsername finds a user by their username.
// It preloads the user's role and the role's permissions.
func (r *MembershipRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	db := r.db.WithContext(ctx)
	err := db.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByIDWithPermissions finds a user by their ID and preloads Role and Permissions.
func (r *MembershipRepository) FindUserByIDWithPermissions(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	db := r.db.WithContext(ctx)
	err := db.Preload("Role.Permissions").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}


// CreateUser creates a new user.
func (r *MembershipRepository) CreateUser(ctx context.Context, user *model.User) error {
	db := r.db.WithContext(ctx)
	return db.Create(user).Error
}

// FindRoleByName finds a role by its name.
func (r *MembershipRepository) FindRoleByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	db := r.db.WithContext(ctx)
	err := db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// CreateRole creates a new role.
func (r *MembershipRepository) CreateRole(ctx context.Context, role *model.Role) error {
	db := r.db.WithContext(ctx)
	return db.Create(role).Error
}

// FindPermissionByName finds a permission by its name.
func (r *MembershipRepository) FindPermissionByName(ctx context.Context, name string) (*model.Permission, error) {
	var permission model.Permission
	db := r.db.WithContext(ctx)
	err := db.Where("name = ?", name).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// CreatePermission creates a new permission.
func (r *MembershipRepository) CreatePermission(ctx context.Context, permission *model.Permission) error {
	db := r.db.WithContext(ctx)
	return db.Create(permission).Error
}

// AssignPermissionsToRole assigns a set of permissions to a role.
func (r *MembershipRepository) AssignPermissionsToRole(ctx context.Context, role *model.Role, permissions []model.Permission) error {
	db := r.db.WithContext(ctx)
	return db.Model(role).Association("Permissions").Replace(permissions)
}

// GetUserPermissions retrieves all permissions for a given user ID.
func (r *MembershipRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var user model.User
	db := r.db.WithContext(ctx)

	if err := db.Preload("Role.Permissions").First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	permissions := make([]string, len(user.Role.Permissions))
	for i, p := range user.Role.Permissions {
		permissions[i] = p.Name
	}

	return permissions, nil
}

// ListUsers retrieves all users.
func (r *MembershipRepository) ListUsers(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	db := r.db.WithContext(ctx)
	err := db.Preload("Role").Find(&users).Error
	return users, err
}

// FindUserByID finds a user by their ID.
func (r *MembershipRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	var user model.User
	db := r.db.WithContext(ctx)
	err := db.Preload("Role").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user's details in the database.
func (r *MembershipRepository) UpdateUser(ctx context.Context, user *model.User) error {
	db := r.db.WithContext(ctx)
	return db.Save(user).Error
}

// FindRoleByID finds a role by its ID.
func (r *MembershipRepository) FindRoleByID(ctx context.Context, roleID uuid.UUID) (*model.Role, error) {
	var role model.Role
	db := r.db.WithContext(ctx)
	err := db.First(&role, "id = ?", roleID).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}
