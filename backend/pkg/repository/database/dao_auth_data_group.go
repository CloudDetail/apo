package database

import (
	"context"
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"gorm.io/gorm"
)

// AuthDataGroup Records the permissions that users or teams have on specific data groups.
type AuthDataGroup struct {
	ID          int64  `gorm:"column:id;primary_key;auto_increment" json:"-"`
	SubjectID   int64  `gorm:"column:subject_id;index:sub_type_id_idx,priority:1" json:"-"`
	SubjectType string `gorm:"column:subject_type;index:sub_type_id_idx,priority:2" json:"-"` // user or team
	DataGroupID int64  `gorm:"column:data_group_id;index:data_group_idx" json:"-"`
	Type        string `gorm:"column:type" json:"type"` // view, edit

	User *User `gorm:"foreignKey:SubjectID;references:UserID" json:"user,omitempty"`
	Team *Team `gorm:"foreignKey:SubjectID;references:TeamID" json:"team,omitempty"`
}

func (AuthDataGroup) TableName() string {
	return "auth_data_group"
}

func (repo *daoRepo) GetAuthDataGroupBySub(subjectID int64, subjectType string) ([]AuthDataGroup, error) {
	var authDataGroups []AuthDataGroup
	err := repo.db.Where("subject_id = ? AND subject_type = ?", subjectID, subjectType).Find(&authDataGroups).Error
	return authDataGroups, err
}

func (repo *daoRepo) AssignDataGroup(ctx context.Context, authDataGroups []AuthDataGroup) error {
	if len(authDataGroups) == 0 {
		return nil
	}
	return repo.GetContextDB(ctx).Save(&authDataGroups).Error
}

func (repo *daoRepo) RevokeDataGroupByGroup(ctx context.Context, dataGroupIDs []int64) error {
	if len(dataGroupIDs) == 0 {
		return nil
	}
	return repo.GetContextDB(ctx).Model(&AuthDataGroup{}).Where("data_group_id IN ?", dataGroupIDs).Delete(nil).Error
}

func (repo *daoRepo) GetModifyAndDeleteDataGroup(subjectID int64, subjectType string, dgPermissions []request.DataGroupPermission) (toModify []AuthDataGroup, toDelete []int64, err error) {
	authGroups, err := repo.GetAuthDataGroupBySub(subjectID, subjectType)
	if err != nil {
		return nil, nil, err
	}

	ids := make([]int64, len(dgPermissions))
	for _, dg := range dgPermissions {
		ids = append(ids, dg.DataGroupID)
	}
	filter := model.DataGroupFilter{
		IDs: ids,
	}
	dataGroups, _, err := repo.GetDataGroup(filter)
	if err != nil {
		return nil, nil, err
	}

	if len(dataGroups) != len(dgPermissions) {
		return nil, nil, model.NewErrWithMessage(errors.New("data group not exist"), code.DataGroupNotExistError)
	}

	hasAuthGroupMap := make(map[int64]AuthDataGroup)
	for _, auth := range authGroups {
		hasAuthGroupMap[auth.DataGroupID] = auth
	}

	for _, dg := range dgPermissions {
		hasDG, ok := hasAuthGroupMap[dg.DataGroupID]
		if !ok {
			if hasDG.Type == dg.PermissionType {
				continue
			}
			authDG := AuthDataGroup{
				DataGroupID: dg.DataGroupID,
				Type:        dg.PermissionType,
				SubjectID:   subjectID,
				SubjectType: subjectType,
			}

			toModify = append(toModify, authDG)
		} else {
			delete(hasAuthGroupMap, dg.DataGroupID)
		}
	}

	for id := range hasAuthGroupMap {
		toDelete = append(toDelete, id)
	}

	return toModify, toDelete, nil
}

func (repo *daoRepo) DeleteAuthDataGroup(ctx context.Context, subjectID int64, subjectType string) error {
	return repo.GetContextDB(ctx).
		Model(&AuthDataGroup{}).
		Where("subject_id = ? AND subject_type = ?", subjectID, subjectType).
		Delete(nil).
		Error
}

func (repo *daoRepo) RevokeDataGroupBySub(ctx context.Context, subjectIDs []int64) error {
	if len(subjectIDs) == 0 {
		return nil
	}
	return repo.GetContextDB(ctx).Model(&AuthDataGroup{}).Where("subject_id IN ?", subjectIDs).Delete(nil).Error
}

func (repo *daoRepo) GetGroupAuthDataGroupByGroup(groupID int64, subjectType string) ([]AuthDataGroup, error) {
	var dataGroups []AuthDataGroup
	err := repo.db.Table("auth_data_group").
		Joins("INNER JOIN data_group dg ON dg.group_id = auth_data_group.data_group_id").
		Where("dg.group_id = ? AND auth_data_group.subject_type = ?", groupID, subjectType).
		Find(&dataGroups).Error
	if err != nil {
		return nil, err
	}
	return dataGroups, nil
}

func (repo *daoRepo) GetDataGroupUsers(groupID int64) ([]AuthDataGroup, error) {
	var ags []AuthDataGroup

	err := repo.db.Preload("User", func(db *gorm.DB) *gorm.DB {
		return db.Select("user_id, username")
	}).
		Select("subject_id, type").
		Where("data_group_id = ? AND subject_type = ?", groupID, model.DATA_GROUP_SUB_TYP_USER).
		Find(&ags).Error
	return ags, err
}

func (repo *daoRepo) GetDataGroupTeams(groupID int64) ([]AuthDataGroup, error) {
	var ags []AuthDataGroup

	err := repo.db.Preload("Team", func(db *gorm.DB) *gorm.DB {
		return db.Select("team_id, team_name")
	}).
		Select("subject_id, type").
		Where("data_group_id = ? AND subject_type = ?", groupID, model.DATA_GROUP_SUB_TYP_TEAM).
		Find(&ags).Error
	return ags, err
}