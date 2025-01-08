// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

// AuthPermission Records which feature are authorised to which subjects.
type AuthPermission struct {
	ID           int    `gorm:"primary_key;auto_increment" json:"id"`
	Type         string `gorm:"column:type;index:idx_sub_id_type;type:varchar(20)" json:"type"` // feature data data group
	SubjectID    int64  `gorm:"column:subject_id;index:idx_sub_id_type" json:"subjectId"`       // Role id, user id or team id.
	SubjectType  string `gorm:"column:subject_type;type:varchar(10)" json:"subjectType"`        // role user team.
	PermissionID int    `gorm:"column:permission_id" json:"permissionId"`
}

func (t *AuthPermission) TableName() string {
	return "auth_permission"
}

func (repo *daoRepo) GetSubjectPermission(subID int64, subType string, typ string) ([]int, error) {
	var permissionIDs []int
	err := repo.db.Model(&AuthPermission{}).
		Select("permission_id").
		Where("subject_id = ? AND subject_type = ? AND type = ?", subID, subType, typ).
		Find(&permissionIDs).Error
	return permissionIDs, err
}

func (repo *daoRepo) GetSubjectsPermission(subIDs []int64, subType string, typ string) ([]AuthPermission, error) {
	var permissions []AuthPermission
	err := repo.db.Model(&AuthPermission{}).
		Where("subject_id in ? AND subject_type = ? AND type = ?", subIDs, subType, typ).
		Find(&permissions).Error
	return permissions, err
}

func (repo *daoRepo) GrantPermission(ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error {
	if len(permissionIDs) == 0 {
		return nil
	}

	db := repo.GetContextDB(ctx)
	permission := make([]AuthPermission, len(permissionIDs))
	for i := range permissionIDs {
		permission[i] = AuthPermission{
			SubjectID:    subID,
			SubjectType:  subType,
			Type:         typ,
			PermissionID: permissionIDs[i],
		}
	}

	return db.Create(&permission).Error
}

// RevokePermission It will delete all record related to sub if permissionIDs is empty.
func (repo *daoRepo) RevokePermission(ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error {
	query := repo.GetContextDB(ctx).
		Model(&AuthPermission{}).
		Where("subject_id = ? AND subject_type = ? AND type = ?", subID, subType, typ)

	if len(permissionIDs) > 0 {
		query = query.Where("permission_id in ?", permissionIDs)
	}

	return query.Delete(nil).Error
}

func (repo *daoRepo) GetAddAndDeletePermissions(subID int64, subType, typ string, permList []int) (toAdd []int, toDelete []int, err error) {
	subPermissions, err := repo.GetSubjectPermission(subID, subType, typ)
	if err != nil {
		return
	}

	permissions, err := repo.GetFeature(nil)

	permissionMap := make(map[int]struct{})
	for _, permission := range permissions {
		permissionMap[permission.FeatureID] = struct{}{}
	}

	subPermissionMap := make(map[int]struct{})
	for _, id := range subPermissions {
		subPermissionMap[id] = struct{}{}
	}

	for _, permission := range permList {
		if _, exists := permissionMap[permission]; !exists {
			err = model.NewErrWithMessage(errors.New("permission does not exist"), code.PermissionNotExistError)
			return
		}
		if _, hasRole := subPermissionMap[permission]; !hasRole {
			toAdd = append(toAdd, permission)
		} else {
			delete(subPermissionMap, permission)
		}
	}

	for permission := range subPermissionMap {
		toDelete = append(toDelete, permission)
	}

	return
}
