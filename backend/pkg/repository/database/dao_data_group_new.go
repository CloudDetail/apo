package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
)

var DaoDataGroup DaoDataGroupNew

type DaoDataGroupNew interface {
	LoadDataGroupTree(ctx core.Context) (*datagroup.DataGroupTreeNode, error)

	GetDataGroupIDsByUserId(ctx core.Context, userID int64) ([]int64, error)
}

func (repo *daoRepo) LoadDataGroupTree(ctx core.Context) (*datagroup.DataGroupTreeNode, error) {
	var res []datagroup.DataGroup
	err := repo.GetContextDB(ctx).Find(&res).Error
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
