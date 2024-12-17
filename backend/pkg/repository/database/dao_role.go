package database

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
)

// Role is a collection of feature permission.
type Role struct {
	RoleID   int    `gorm:"column:role_id;primary_key" json:"roleId"`
	RoleName string `gorm:"column:role_name;uniqueIndex" json:"roleName"`
}

type UserRole struct {
	UserID int64 `gorm:"column:user_id;primary_key;"`
	RoleID int   `gorm:"column:role_id;primary_key;"`
}

func (t *Role) TableName() string {
	return "role"
}

func (t *UserRole) TableName() string {
	return "user_role"
}

// GetRoles Get all roles for the given condition.
func (repo *daoRepo) GetRoles(filter model.RoleFilter) ([]Role, error) {
	var roles []Role
	query := repo.db

	if len(filter.Names) > 0 {
		query = query.Where("role_name in ?", filter.Names)
	}
	if len(filter.Name) > 0 {
		query = query.Where("role_name = ?", filter.Name)
	}
	if len(filter.IDs) > 0 {
		query = query.Where("role_id in ?", filter.IDs)
	}
	if filter.ID != 0 {
		query = query.Where("role_id = ?", filter.ID)
	}

	err := query.Find(&roles).Error
	return roles, err
}

// GetUserRole Get user's role.
func (repo *daoRepo) GetUserRole(userID int64) ([]UserRole, error) {
	var userRoles []UserRole
	err := repo.db.Where("user_id = ?", userID).Find(&userRoles).Error
	return userRoles, err
}

// GetUsersRole Get user's role in batch.
func (repo *daoRepo) GetUsersRole(userIDs []int64) ([]UserRole, error) {
	var userRoles []UserRole
	err := repo.db.Where("user_id in ?", userIDs).Error
	return userRoles, err
}

func (repo *daoRepo) RoleGranted(userID int64, roleID int) (bool, error) {
	var count int64
	if err := repo.db.Model(&UserRole{}).Where("role_id = ? AND user_id = ?", roleID, userID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *daoRepo) GrantRole(ctx context.Context, userID int64, roleIDs []int) error {
	db := repo.GetContextDB(ctx)
	userRole := make([]UserRole, len(roleIDs))
	for i, roleID := range roleIDs {
		userRole[i] = UserRole{
			UserID: userID,
			RoleID: roleID,
		}
	}

	return db.Create(&userRole).Error
}

func (repo *daoRepo) RevokeRole(ctx context.Context, userID int64, roleIDs []int) error {
	return repo.GetContextDB(ctx).
		Model(&UserRole{}).Where("user_id = ? AND role_id in ?", userID, roleIDs).Delete(nil).Error
}

func (repo *daoRepo) RoleExists(roleID int64) (bool, error) {
	var count int64
	if err := repo.db.Model(&Role{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}
