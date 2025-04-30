// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"errors"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Role is a collection of feature permission.
type Role struct {
	RoleID      int    `gorm:"column:role_id;primary_key;auto_increment" json:"roleId"`
	RoleName    string `gorm:"column:role_name;type:varchar(20);uniqueIndex" json:"roleName"`
	Description string `gorm:"column:description;type:varchar(50)" json:"description"`
}

type UserRole struct {
	UserID int64 `gorm:"column:user_id;primary_key"`
	RoleID int   `gorm:"column:role_id;primary_key"`
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
	query := repo.db.Where("role_name != ?", AnonymousUsername)

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
	err := repo.db.Where("user_id in ?", userIDs).Find(&userRoles).Error
	return userRoles, err
}

func (repo *daoRepo) RoleGrantedToUser(userID int64, roleID int) (bool, error) {
	var count int64
	if err := repo.db.Model(&UserRole{}).Where("role_id = ? AND user_id = ?", roleID, userID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *daoRepo) GrantRoleWithUser(ctx context.Context, userID int64, roleIDs []int) error {
	if len(roleIDs) == 0 {
		return nil
	}
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

func (repo *daoRepo) RoleExists(roleID int) (bool, error) {
	var count int64
	if err := repo.db.Model(&Role{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}

func (repo *daoRepo) CreateRole(ctx context.Context, role *Role) error {
	return repo.GetContextDB(ctx).Create(&role).Error
}

func (repo *daoRepo) DeleteRole(ctx context.Context, roleID int) error {
	err := repo.GetContextDB(ctx).Model(&Role{}).Where("role_id = ?", roleID).Delete(nil).Error
	if err != nil {
		return err
	}

	return repo.GetContextDB(ctx).
		Model(&UserRole{}).
		Where("role_id = ?", roleID).
		Delete(nil).Error
}

func (repo *daoRepo) UpdateRole(ctx context.Context, roleID int, roleName, description string) error {
	role := Role{
		RoleID:      roleID,
		RoleName:    roleName,
		Description: description,
	}

	return repo.GetContextDB(ctx).Updates(&role).Error
}

func (repo *daoRepo) GrantRoleWithRole(ctx context.Context, roleID int, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	userRoles := make([]UserRole, len(userIDs))
	for i, userID := range userIDs {
		userRoles[i] = UserRole{
			UserID: userID,
			RoleID: roleID,
		}
	}

	return repo.GetContextDB(ctx).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
			DoNothing: true,
		}).Create(&userRoles).Error
}

func (repo *daoRepo) RevokeRoleWithRole(ctx context.Context, roleID int) error {
	return repo.GetContextDB(ctx).
		Model(&UserRole{}).Where("role_id = ?", roleID).Delete(nil).Error
}

func (repo *daoRepo) RoleGranted(roleID int) (bool, error) {
	var count int64

	err := repo.db.Model(&UserRole{}).Where("role_id = ?", roleID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
