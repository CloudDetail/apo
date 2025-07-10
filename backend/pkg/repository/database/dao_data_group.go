// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"gorm.io/gorm/clause"
)

// DatasourceGroup is a mapping table of Datasource and DataGroup.
type DatasourceGroup struct {
	GroupID    int64  `gorm:"column:group_id;primary_key" json:"-"`
	Datasource string `gorm:"column:datasource;primary_key" json:"datasource"`
	Type       string `gorm:"column:type" json:"type"`         // service or namespace
	Category   string `gorm:"column:category" json:"category"` // apm or normal

	ClusterID string `gorm:"column:cluster_id" json:"clusterId"`
}

func (dsg *DatasourceGroup) TableName() string {
	return "datasource_group"
}

func (repo *daoRepo) CreateDataGroup(ctx core.Context, group *datagroup.DataGroup) error {
	return repo.GetContextDB(ctx).Create(group).Error
}

func (repo *daoRepo) DeleteDataGroup(ctx core.Context, groupID int64) error {
	group := datagroup.DataGroup{
		GroupID: groupID,
	}
	if err := repo.GetContextDB(ctx).Delete(&group).Error; err != nil {
		return err
	}

	return repo.GetContextDB(ctx).Model(&AuthDataGroup{}).Where("data_group_id = ?", groupID).Delete(nil).Error
}

func (repo *daoRepo) CreateDatasourceGroup(ctx core.Context, datasource []model.Datasource, dataGroupID int64) error {
	if len(datasource) == 0 {
		return nil
	}

	datasourceGroups := make([]DatasourceGroup, 0, len(datasource))
	for _, ds := range datasource {
		dsGroup := DatasourceGroup{
			GroupID:    dataGroupID,
			Datasource: ds.Datasource,
			Type:       ds.Type,
			Category:   ds.Category,
		}
		datasourceGroups = append(datasourceGroups, dsGroup)
	}

	return repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "datasource"}},
		DoNothing: true,
	}).Create(&datasourceGroups).Error
}

func (repo *daoRepo) DeleteDSGroup(ctx core.Context, groupID int64) error {
	return repo.GetContextDB(ctx).Model(&DatasourceGroup{}).Where("group_id = ?", groupID).Delete(&DatasourceGroup{}).Error
}

func (repo *daoRepo) UpdateDataGroup(ctx core.Context, groupID int64, groupName string, description string) error {
	return repo.GetContextDB(ctx).
		Model(&datagroup.DataGroup{}).
		Where("group_id = ?", groupID).
		Update("group_name", groupName).
		Update("description", description).
		Error
}

// DataGroupExist check whether the group exists for the given condition.
func (repo *daoRepo) DataGroupExist(ctx core.Context, filter model.DataGroupFilter) (bool, error) {
	var count int64
	query := repo.GetContextDB(ctx)
	if len(filter.Name) > 0 {
		query = query.Where("group_name = ?", filter.Name)
	}
	if filter.ID > 0 {
		query = query.Where("group_id = ?", filter.ID)
	}
	if len(filter.IDs) > 0 {
		query = query.Where("group_id IN ?", filter.IDs)
	}
	if err := query.Model(&datagroup.DataGroup{}).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *daoRepo) GetDataGroup(ctx core.Context, filter model.DataGroupFilter) ([]datagroup.DataGroup, int64, error) {
	var (
		dataGroups []datagroup.DataGroup
		count      int64
	)
	query := repo.GetContextDB(ctx)
	if len(filter.Name) > 0 {
		query = query.Where("group_name like ?", "%"+filter.Name+"%")
	}
	if len(filter.Names) > 0 {
		query = query.Where("group_name in ?", filter.Names)
	}
	if filter.ID != 0 {
		query = query.Where("group_id = ?", filter.ID)
	}
	if len(filter.IDs) > 0 {
		query = query.Where("group_id in ?", filter.IDs)
	}
	if len(filter.DatasourceList) > 0 {
		conditions := make([][]interface{}, 0, len(filter.DatasourceList))
		for _, item := range filter.DatasourceList {
			conditions = append(conditions, []interface{}{item.Datasource, item.Type})
		}

		subQuery := repo.GetContextDB(ctx).Model(&DatasourceGroup{}).
			Select("group_id").
			Where("(datasource, type) IN ?", conditions)

		query.Where("group_id IN (?)", subQuery)
	}

	if err := query.Model(&datagroup.DataGroup{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if filter.CurrentPage != nil && filter.PageSize != nil {
		query = query.Offset((*filter.CurrentPage - 1) * (*filter.PageSize)).Limit(*filter.PageSize)
	}

	err := query.Find(&dataGroups).Error
	return dataGroups, count, err
}

func (repo *daoRepo) RetrieveDataFromGroup(ctx core.Context, groupID int64, datasource []string) error {
	return repo.GetContextDB(ctx).
		Model(&DatasourceGroup{}).Where("group_id = ? AND datasource in ?", groupID, datasource).Delete(nil).Error
}

func (repo *daoRepo) GetGroupDatasource(ctx core.Context, groupID ...int64) ([]DatasourceGroup, error) {
	var dsGroup []DatasourceGroup
	err := repo.GetContextDB(ctx).Where("group_id in ?", groupID).Find(&dsGroup).Error
	return dsGroup, err
}

func (repo *daoRepo) GetSubjectDataGroupList(ctx core.Context, subjectID int64, subjectType string) ([]datagroup.DataGroup, error) {
	var dataGroups []datagroup.DataGroup

	err := repo.GetContextDB(ctx).Table("data_group").
		Select("data_group.group_id, data_group.group_name, data_group.description, auth_data_group.type as auth_type").
		Joins("JOIN auth_data_group ON auth_data_group.data_group_id = data_group.group_id").
		Where("auth_data_group.subject_id = ? AND auth_data_group.subject_type = ?", subjectID, subjectType).
		Find(&dataGroups).Error

	if err != nil {
		return nil, err
	}

	return dataGroups, nil
}

func (repo *daoRepo) GetDataGroupByGroupIDOrUserID(ctx core.Context, groupID int64, userID int64, category string) ([]datagroup.DataGroup, error) {
	if groupID != 0 {
		return repo.GetDataGroupByGroupID(ctx, groupID, category)
	}
	return repo.GetDataGroupByUserID(ctx, userID)
}

func (repo *daoRepo) GetDataGroupByGroupID(ctx core.Context, groupID int64, category string) ([]datagroup.DataGroup, error) {
	if groupID == 0 {
		return nil, fmt.Errorf("group id is empty")
	}

	filter := model.DataGroupFilter{
		ID: groupID,
	}

	dataGroups, _, err := repo.GetDataGroup(ctx, filter)
	if err != nil {
		return dataGroups, err
	}

	if len(dataGroups) == 0 {
		return nil, core.Error(code.DataGroupNotExistError, "data group does not exits")
	}
	return dataGroups, nil
}

func (repo *daoRepo) GetDataGroupByUserID(ctx core.Context, userID int64) ([]datagroup.DataGroup, error) {
	teamIDs, err := repo.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, err
	}

	seen := make(map[int64]struct{})
	// Get user's teams.
	var groups []datagroup.DataGroup
	for _, teamID := range teamIDs {
		gs, err := repo.GetSubjectDataGroupList(ctx, teamID, model.DATA_GROUP_SUB_TYP_TEAM)
		if err != nil {
			return nil, err
		}

		for _, g := range gs {
			if _, ok := seen[g.GroupID]; ok {
				continue
			}

			seen[g.GroupID] = struct{}{}
			groups = append(groups, g)
		}
	}

	for i := range groups {
		groups[i].Source = model.DATA_GROUP_SUB_TYP_TEAM
	}

	gs, err := repo.GetSubjectDataGroupList(ctx, userID, model.DATA_GROUP_SUB_TYP_USER)
	for i := range gs {
		gs[i].Source = model.DATA_GROUP_SUB_TYP_USER
	}

	if err != nil {
		return nil, err
	}

	groups = append(groups, gs...)

	return groups, nil
}
