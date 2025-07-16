// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var DaoDataGroup DaoDataGroupNew

type DaoDataGroupNew interface {
	LoadDataGroupTree(ctx core.Context) (*datagroup.DataGroupTreeNode, error)

	GetDataGroupIDsByUserId(ctx core.Context, userID int64) ([]int64, error)

	InitRootGroup(ctx core.Context) error
}

func (repo *daoRepo) LoadDataGroupTree(ctx core.Context) (*datagroup.DataGroupTreeNode, error) {
	var res []datagroup.DataGroup
	err := repo.GetContextDB(ctx).Find(&res).Order("group_id ASC").Error
	if err != nil {
		return nil, err
	}

	var root *datagroup.DataGroupTreeNode
	var nodesMap = make(map[int64]*datagroup.DataGroupTreeNode)
	for i := 0; i < len(res); i++ {
		treeNode := &datagroup.DataGroupTreeNode{
			DataGroup: res[i],
			SubGroups: []*datagroup.DataGroupTreeNode{},
		}
		if res[i].GroupID == 0 {
			root = treeNode
		}
		nodesMap[res[i].GroupID] = treeNode
	}

	if root == nil {
		root = &datagroup.DataGroupTreeNode{
			DataGroup: datagroup.DataGroup{
				GroupID:       0,
				GroupName:     "ALL",
				Description:   "Contains all data",
				ParentGroupID: -1,
			},
			SubGroups: []*datagroup.DataGroupTreeNode{},
		}
	}

	for _, node := range nodesMap {
		if node.GroupID == 0 {
			continue
		}
		if node.ParentGroupID == 0 {
			root.SubGroups = append(root.SubGroups, node)
			continue
		}
		if parentNode, ok := nodesMap[node.ParentGroupID]; ok {
			parentNode.SubGroups = append(parentNode.SubGroups, node)
		}
	}

	root.RecursiveSortSubGroups()
	return root, nil
}

func (repo *daoRepo) GetDataGroupIDsByUserId(ctx core.Context, userID int64) ([]int64, error) {
	teamIDs, err := repo.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user's teams.
	var groupIDs = make(map[int64]struct{})
	for _, teamID := range teamIDs {
		gs, err := repo.GetSubjectDataGroupList(ctx, teamID, model.DATA_GROUP_SUB_TYP_TEAM)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(gs); i++ {
			groupIDs[gs[i].GroupID] = struct{}{}
		}
	}

	gs, err := repo.GetSubjectDataGroupList(ctx, userID, model.DATA_GROUP_SUB_TYP_USER)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(gs); i++ {
		groupIDs[gs[i].GroupID] = struct{}{}
	}

	var groups []int64
	for groupID := range groupIDs {
		groups = append(groups, groupID)
	}
	return groups, nil
}

func (repo *daoRepo) assignUserDataGroupIfNotExist(ctx core.Context, userID, groupID int64) error {
	var count int64
	err := repo.GetContextDB(ctx).Model(&AuthDataGroup{}).
		Where("subject_id = ? AND type = ? AND data_group_id = ?", userID, model.DATA_GROUP_SUB_TYP_USER, groupID).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	return repo.GetContextDB(ctx).Create(&AuthDataGroup{
		SubjectID:   userID,
		SubjectType: model.DATA_GROUP_SUB_TYP_USER,
		GroupID:     groupID,
		Type:        "view",
	}).Error
}

func (repo *daoRepo) InitRootGroup(ctx core.Context) error {
	var count int64
	err := repo.GetContextDB(ctx).
		Model(&datagroup.DataGroup{}).
		Where("group_id = ?", 0).
		Count(&count).Error

	if err != nil {
		return err
	}

	if count == 0 {
		// using Exec to avoid auto increment of group_id
		err := repo.GetContextDB(ctx).Exec("INSERT INTO data_group (group_id, group_name, description, parent_group_id) VALUES (0, 'ALL', 'Contains all data', -1)").Error
		if err != nil {
			return err
		}
	}

	if anonymousUser, err := repo.GetAnonymousUser(ctx); err == nil {
		err := repo.assignUserDataGroupIfNotExist(ctx, anonymousUser.UserID, 0)
		if err != nil {
			return err
		}
	}

	if adminUser, err := repo.GetAdminUser(ctx); err == nil {
		err := repo.assignUserDataGroupIfNotExist(ctx, adminUser.UserID, 0)
		if err != nil {
			return err
		}
	}

	// migrate-datasourceGroup
	err = repo.GetContextDB(ctx).Model(&datagroup.DataGroup2Scope{}).Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		var group2Scope []datagroup.DataGroup2Scope
		err := repo.GetContextDB(ctx).Table("data_scope ds").
			Select("group_id", "scope_id").
			Joins("INNER JOIN datasource_group dsg ON (dsg.datasource = ds.namespace and dsg.type = 'namespace') or (dsg.datasource = ds.service and dsg.type = 'service')").
			Group("dsg.group_id,ds.scope_id").
			Find(&group2Scope).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if len(group2Scope) > 0 {
			err = repo.GetContextDB(ctx).Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&group2Scope).Error

			if err != nil {
				return err
			}
		}
	}

	err = repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&datagroup.DataScope{
		ScopeID:  "APO_ALL_DATA",
		Category: "system",
		Name:     "ALL",
		Type:     "system",
	}).Error

	if err != nil {
		return err
	}

	err = repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&datagroup.DataGroup2Scope{
		GroupID: 0,
		ScopeID: "APO_ALL_DATA",
	}).Error

	return err
}
