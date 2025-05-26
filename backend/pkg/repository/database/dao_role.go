// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// GetRoles Get all roles for the given condition.
func (repo *daoRepo) GetRoles(ctx core.Context, filter model.RoleFilter) ([]profile.Role, error) {
	var roles []profile.Role
	query := repo.GetContextDB(ctx).Where("role_name != ?", AnonymousUsername)

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
func (repo *daoRepo) GetUserRole(ctx core.Context, userID int64) ([]profile.UserRole, error) {
	var userRoles []profile.UserRole
	err := repo.GetContextDB(ctx).Where("user_id = ?", userID).Find(&userRoles).Error
	return userRoles, err
}

// GetUsersRole Get user's role in batch.
func (repo *daoRepo) GetUsersRole(ctx core.Context, userIDs []int64) ([]profile.UserRole, error) {
	var userRoles []profile.UserRole
	err := repo.GetContextDB(ctx).Where("user_id in ?", userIDs).Find(&userRoles).Error
	return userRoles, err
}

func (repo *daoRepo) RoleGrantedToUser(ctx core.Context, userID int64, roleID int) (bool, error) {
	var count int64
	if err := repo.GetContextDB(ctx).Model(&profile.UserRole{}).Where("role_id = ? AND user_id = ?", roleID, userID).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *daoRepo) GrantRoleWithUser(ctx core.Context, userID int64, roleIDs []int) error {
	if len(roleIDs) == 0 {
		return nil
	}
	db := repo.GetContextDB(ctx)
	userRole := make([]profile.UserRole, len(roleIDs))
	for i, roleID := range roleIDs {
		userRole[i] = profile.UserRole{
			UserID: userID,
			RoleID: roleID,
		}
	}

	return db.Create(&userRole).Error
}

func (repo *daoRepo) RevokeRole(ctx core.Context, userID int64, roleIDs []int) error {
	return repo.GetContextDB(ctx).
		Model(&profile.UserRole{}).Where("user_id = ? AND role_id in ?", userID, roleIDs).Delete(nil).Error
}

func (repo *daoRepo) RoleExists(ctx core.Context, roleID int) (bool, error) {
	var count int64
	if err := repo.GetContextDB(ctx).Model(&profile.Role{}).Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}

func (repo *daoRepo) CreateRole(ctx core.Context, role *profile.Role) error {
	return repo.GetContextDB(ctx).Create(&role).Error
}

func (repo *daoRepo) DeleteRole(ctx core.Context, roleID int) error {
	err := repo.GetContextDB(ctx).Model(&profile.Role{}).Where("role_id = ?", roleID).Delete(nil).Error
	if err != nil {
		return err
	}

	return repo.GetContextDB(ctx).
		Model(&profile.UserRole{}).
		Where("role_id = ?", roleID).
		Delete(nil).Error
}

func (repo *daoRepo) UpdateRole(ctx core.Context, roleID int, roleName, description string) error {
	role := profile.Role{
		RoleID:      roleID,
		RoleName:    roleName,
		Description: description,
	}

	return repo.GetContextDB(ctx).Updates(&role).Error
}

func (repo *daoRepo) GrantRoleWithRole(ctx core.Context, roleID int, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}

	userRoles := make([]profile.UserRole, len(userIDs))
	for i, userID := range userIDs {
		userRoles[i] = profile.UserRole{
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

func (repo *daoRepo) RevokeRoleWithRole(ctx core.Context, roleID int) error {
	return repo.GetContextDB(ctx).
		Model(&profile.UserRole{}).Where("role_id = ?", roleID).Delete(nil).Error
}

func (repo *daoRepo) RoleGranted(ctx core.Context, roleID int) (bool, error) {
	var count int64

	err := repo.GetContextDB(ctx).Model(&profile.UserRole{}).Where("role_id = ?", roleID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// userID -> roles
type userRoleMap map[int64][]profile.Role

func (m userRoleMap) Get(userID int64) []profile.Role {
	if m == nil {
		return []profile.Role{}
	}
	if v, find := m[userID]; find {
		return v
	}
	return []profile.Role{}
}

type userRole struct {
	UserID int64 `gorm:"column:user_id"`
	profile.Role
}

func (repo *daoRepo) getRoleByUserID(ctx core.Context, userID ...int64) (userRoleMap, error) {
	var userRoles []userRole
	err := repo.GetContextDB(ctx).Model(&profile.UserRole{}).
		Select("user_id", "role.role_id", "role_name", "description").
		Where("user_id IN ?", userID).
		Joins("LEFT JOIN role ON user_role.role_id = role.role_id").
		Order("user_id").
		Scan(&userRoles).Error

	if err != nil {
		return nil, err
	}

	userRoleMap := make(userRoleMap)
	for _, userRole := range userRoles {
		userRoleMap[userRole.UserID] = append(userRoleMap[userRole.UserID], userRole.Role)
	}
	return userRoleMap, nil
}
