// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
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

func (repo *daoRepo) InitRootGroup(ctx core.Context) error {
	err := repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&datagroup.DataGroup{
		GroupID:       0,
		GroupName:     "ALL",
		Description:   "Contains all data",
		ParentGroupID: -1,
	}).Error

	if err != nil {
		return err
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

	// migrate-datasourceGroup
	var count int64
	err = repo.GetContextDB(ctx).Model(&datagroup.DataGroup2Scope{}).Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		err = repo.GetContextDB(ctx).Clauses(clause.OnConflict{
			DoNothing: true,
		}).Select("ds.group_id, ds.scope_id").
			Table("data_scope AS ds").
			Joins(`INNER JOIN datasource_group AS dsg ON (dsg.datasource = ds.namespace AND dsg.type = 'namespace') OR (dsg.datasource = ds.service AND dsg.type = 'service')`).
			Group("ds.group_id, ds.scope_id").
			Create(nil).Error

		if err != nil {
			return err
		}
	}

	err = repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&datagroup.DataGroup2Scope{
		GroupID: 0,
		ScopeID: "APO_ALL_DATA",
	}).Error

	return err
}
