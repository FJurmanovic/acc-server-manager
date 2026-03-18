package repository

import (
	"acc-server-manager/local/model"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MembershipRepository struct {
	*BaseRepository[model.User, model.MembershipFilter]
}

func NewMembershipRepository(db *gorm.DB) *MembershipRepository {
	return &MembershipRepository{
		BaseRepository: NewBaseRepository[model.User, model.MembershipFilter](db, model.User{}),
	}
}

func (r *MembershipRepository) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	db := r.db.WithContext(ctx)
	err := db.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MembershipRepository) FindUserByIDWithPermissions(ctx context.Context, userID string) (*model.User, error) {
	var user model.User
	db := r.db.WithContext(ctx)
	err := db.Preload("Role.Permissions").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MembershipRepository) CreateUser(ctx context.Context, user *model.User) error {
	db := r.db.WithContext(ctx)
	return db.Create(user).Error
}

func (r *MembershipRepository) FindRoleByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	db := r.db.WithContext(ctx)
	err := db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *MembershipRepository) CreateRole(ctx context.Context, role *model.Role) error {
	db := r.db.WithContext(ctx)
	return db.Create(role).Error
}

func (r *MembershipRepository) FindPermissionByName(ctx context.Context, name string) (*model.Permission, error) {
	var permission model.Permission
	db := r.db.WithContext(ctx)
	err := db.Where("name = ?", name).First(&permission).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *MembershipRepository) CreatePermission(ctx context.Context, permission *model.Permission) error {
	db := r.db.WithContext(ctx)
	return db.Create(permission).Error
}

func (r *MembershipRepository) AssignPermissionsToRole(ctx context.Context, role *model.Role, permissions []model.Permission) error {
	db := r.db.WithContext(ctx)
	return db.Model(role).Association("Permissions").Replace(permissions)
}

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

func (r *MembershipRepository) ListUsers(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	db := r.db.WithContext(ctx)
	err := db.Preload("Role").Find(&users).Error
	return users, err
}

func (r *MembershipRepository) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	db := r.db.WithContext(ctx)
	return db.Delete(&model.User{}, "id = ?", userID).Error
}

func (r *MembershipRepository) FindUserByID(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	var user model.User
	db := r.db.WithContext(ctx)
	err := db.Preload("Role").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *MembershipRepository) UpdateUser(ctx context.Context, user *model.User) error {
	db := r.db.WithContext(ctx)
	return db.Save(user).Error
}

func (r *MembershipRepository) FindRoleByID(ctx context.Context, roleID uuid.UUID) (*model.Role, error) {
	var role model.Role
	db := r.db.WithContext(ctx)
	err := db.First(&role, "id = ?", roleID).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *MembershipRepository) ListUsersWithFilter(ctx context.Context, filter *model.MembershipFilter) (*[]model.User, error) {
	return r.BaseRepository.GetAll(ctx, filter)
}

func (r *MembershipRepository) ListRoles(ctx context.Context) ([]*model.Role, error) {
	var roles []*model.Role
	db := r.db.WithContext(ctx)
	err := db.Find(&roles).Error
	return roles, err
}
