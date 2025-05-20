// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"encoding/json"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"gorm.io/gorm"
)

// AuthDataGroup Records the permissions that users or teams have on specific data groups.
type AuthDataGroup struct {
	ID          int64  `gorm:"column:id;primary_key;auto_increment" json:"-"`
	SubjectID   int64  `gorm:"column:subject_id;index:sub_type_id_idx,priority:1" json:"-"`
	SubjectType string `gorm:"column:subject_type;index:sub_type_id_idx,priority:2" json:"-"` // user or team
	GroupID     int64  `gorm:"column:data_group_id;index:group_id_idx" json:"-"`
	Type        string `gorm:"column:type;default:view" json:"type"` // view, edit

	User *User `gorm:"-" json:"user,omitempty"`
	Team *Team `gorm:"-" json:"team,omitempty"`
}

func (adg AuthDataGroup) MarshalJSON() ([]byte, error) {
	result := map[string]interface{}{
		"type": adg.Type,
	}

	if adg.User != nil {
		userData, err := json.Marshal(adg.User)
		if err != nil {
			return nil, err
		}
		var userMap map[string]interface{}
		err = json.Unmarshal(userData, &userMap)
		if err != nil {
			return nil, err
		}
		for k, v := range userMap {
			result[k] = v
		}
	}

	if adg.Team != nil {
		teamData, err := json.Marshal(adg.Team)
		if err != nil {
			return nil, err
		}
		var teamMap map[string]interface{}
		err = json.Unmarshal(teamData, &teamMap)
		if err != nil {
			return nil, err
		}
		for k, v := range teamMap {
			result[k] = v
		}
	}

	return json.Marshal(result)
}

func (AuthDataGroup) TableName() string {
	return "auth_data_group"
}

func (repo *daoRepo) GetAuthDataGroupBySub(ctx core.Context, subjectID int64, subjectType string) ([]AuthDataGroup, error) {
	var authDataGroups []AuthDataGroup
	err := repo.GetContextDB(ctx).Where("subject_id = ? AND subject_type = ?", subjectID, subjectType).Find(&authDataGroups).Error
	return authDataGroups, err
}

func (repo *daoRepo) AssignDataGroup(ctx core.Context, authDataGroups []AuthDataGroup) error {
	if len(authDataGroups) == 0 {
		return nil
	}
	return repo.GetContextDB(ctx).Save(&authDataGroups).Error
}

func (repo *daoRepo) RevokeDataGroupByGroup(ctx core.Context, dataGroupIDs []int64, subjectID int64) error {
	if len(dataGroupIDs) == 0 {
		return nil
	}
	return repo.GetContextDB(ctx).Model(&AuthDataGroup{}).Where("data_group_id IN ? AND subject_id = ?", dataGroupIDs, subjectID).Delete(nil).Error
}

func (repo *daoRepo) GetModifyAndDeleteDataGroup(ctx core.Context, subjectID int64, subjectType string, dgPermissions []request.DataGroupPermission) (toModify []AuthDataGroup, toDelete []int64, err error) {
	authGroups, err := repo.GetAuthDataGroupBySub(ctx, subjectID, subjectType)
	if err != nil {
		return nil, nil, err
	}

	if len(dgPermissions) > 0 {
		ids := make([]int64, len(dgPermissions))
		for _, dg := range dgPermissions {
			ids = append(ids, dg.DataGroupID)
		}
		filter := model.DataGroupFilter{
			IDs: ids,
		}
		exists, err := repo.DataGroupExist(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
		if !exists {
			return nil, nil, core.Error(code.DataGroupNotExistError, "data group not exist")
		}
	}

	hasAuthGroupMap := make(map[int64]AuthDataGroup)
	for _, auth := range authGroups {
		hasAuthGroupMap[auth.GroupID] = auth
	}

	for _, dg := range dgPermissions {
		hasDG, ok := hasAuthGroupMap[dg.DataGroupID]
		if !ok {
			if hasDG.Type == dg.PermissionType {
				continue
			}
			authDG := AuthDataGroup{
				GroupID:     dg.DataGroupID,
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

func (repo *daoRepo) DeleteAuthDataGroup(ctx core.Context, subjectID int64, subjectType string) error {
	return repo.GetContextDB(ctx).
		Model(&AuthDataGroup{}).
		Where("subject_id = ? AND subject_type = ?", subjectID, subjectType).
		Delete(nil).
		Error
}

func (repo *daoRepo) RevokeDataGroupBySub(ctx core.Context, subjectIDs []int64, groupID int64) error {
	if len(subjectIDs) == 0 {
		return nil
	}
	return repo.GetContextDB(ctx).Model(&AuthDataGroup{}).Where("subject_id IN ? AND data_group_id = ?", subjectIDs, groupID).Delete(nil).Error
}

func (repo *daoRepo) GetGroupAuthDataGroupByGroup(ctx core.Context, groupID int64, subjectType string) ([]AuthDataGroup, error) {
	var dataGroups []AuthDataGroup
	err := repo.GetContextDB(ctx).Table("auth_data_group").
		Joins("INNER JOIN data_group dg ON dg.group_id = auth_data_group.data_group_id").
		Where("dg.group_id = ? AND auth_data_group.subject_type = ?", groupID, subjectType).
		Find(&dataGroups).Error
	if err != nil {
		return nil, err
	}
	return dataGroups, nil
}

func (repo *daoRepo) GetDataGroupUsers(ctx core.Context, groupID int64) ([]AuthDataGroup, error) {
	var ags []AuthDataGroup
	err := repo.GetContextDB(ctx).
		Select("subject_id", "type").
		Where("data_group_id = ? AND subject_type = ?", groupID, model.DATA_GROUP_SUB_TYP_USER).
		Find(&ags).Error
	if err != nil {
		return nil, err
	}

	if len(ags) == 0 {
		return ags, nil
	}

	for i := 0; i < len(ags); i++ {
		var user User
		err = repo.GetContextDB(ctx).
			Select("user_id", "username").
			First(&user, "user_id = ?", ags[i].SubjectID).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		if err == nil {
			ags[i].User = &user
		}
	}

	return ags, nil
}

func (repo *daoRepo) GetDataGroupTeams(ctx core.Context, groupID int64) ([]AuthDataGroup, error) {
	var ags []AuthDataGroup

	err := repo.GetContextDB(ctx).
		Select("subject_id", "type").
		Where("data_group_id = ? AND subject_type = ?", groupID, model.DATA_GROUP_SUB_TYP_TEAM).
		Find(&ags).Error
	if err != nil {
		return nil, err
	}

	if len(ags) == 0 {
		return ags, nil
	}

	for i := 0; i < len(ags); i++ {
		var team Team
		err = repo.GetContextDB(ctx).
			Select("team_id", "team_name").
			First(&team, "team_id = ?", ags[i].SubjectID).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, err
		}

		if err == nil {
			ags[i].Team = &team
		}
	}

	return ags, err
}

func (repo *daoRepo) CheckGroupPermission(ctx core.Context, userID, groupID int64, typ string) (bool, error) {
	var (
		count int64
		err   error
	)

	query := repo.GetContextDB(ctx).Model(&AuthDataGroup{}).
		Where("subject_id = ? AND data_group_id = ? AND subject_type = ?", userID, groupID, model.DATA_GROUP_SUB_TYP_USER)

	if typ == "edit" {
		query = query.Where(`"type" = ?`, typ)
	}

	err = query.Count(&count).Error
	if err != nil {
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	teamIDs, err := repo.GetUserTeams(ctx, userID)
	if err != nil {
		return false, err
	}

	query = repo.GetContextDB(ctx).Model(&AuthDataGroup{}).
		Where("subject_id IN ? AND data_group_id = ? AND subject_type = ?", teamIDs, groupID, model.DATA_GROUP_SUB_TYP_TEAM)

	if typ == "edit" {
		query = query.Where(`"type" = ?`, typ)
	}

	err = query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
