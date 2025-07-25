// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"slices"
	"sync"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var (
	DataGroupStorage *DataGroupStore
	once             sync.Once
)

func CutTopologyRelationInGroup(ctx core.Context, dbRepo database.Repo, groupID int64, topologyRelation []*model.TopologyRelation) ([]*model.TopologyRelation, error) {
	if groupID == 0 {
		return topologyRelation, nil
	}

	selected, err := dbRepo.GetScopeIDsSelectedByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	svcList := DataGroupStorage.GetFullPermissionSvcList(selected)
	cutRelation := make([]*model.TopologyRelation, 0)
	for _, relation := range topologyRelation {
		if relation.Group != model.GROUP_SERVICE ||
			slices.Contains(svcList, relation.Service) ||
			slices.Contains(svcList, relation.ParentService) {
			cutRelation = append(cutRelation, relation)
		}
	}
	return cutRelation, nil
}

func MarkTopologyNodeInGroup(ctx core.Context, dbRepo database.Repo, groupID int64, topologyNode *model.TopologyNodes) (*model.TopologyNodes, error) {
	if groupID == 0 {
		return topologyNode, nil
	}

	selected, err := dbRepo.GetScopeIDsSelectedByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	svcList := DataGroupStorage.GetFullPermissionSvcList(selected)

	for _, node := range topologyNode.Nodes {
		if node.Group == model.GROUP_SERVICE && !slices.Contains(svcList, node.Service) {
			node.OutOfGroup = true
		}
	}
	return topologyNode, nil
}

func GetPQLFilterByGroupID(ctx core.Context, dbRepo database.Repo, category string, groupID int64) (prometheus.PQLFilter, error) {
	if groupID == 0 {
		return prometheus.NewFilter(), nil
	}
	scopes, err := dbRepo.GetScopesByGroupIDAndCat(ctx, groupID, category)
	if err != nil {
		return prometheus.NewFilter(), err
	}
	scopeTree := convertScopesToScopeTree(scopes)
	return ConvertScopeNodeToPQLFilter(scopeTree), nil
}

func ConvertScopeNodeToPQLFilter(scopeNode *datagroup.DataScopeTreeNode) prometheus.PQLFilter {
	if scopeNode == nil {
		return prometheus.AlwaysFalseFilter
	}

	if scopeNode.Type == datagroup.DATASOURCE_TYP_SYSTEM && scopeNode.IsChecked {
		return nil
	}

	var filters []prometheus.PQLFilter

	var dfs func(node *datagroup.DataScopeTreeNode) prometheus.PQLFilter
	dfs = func(node *datagroup.DataScopeTreeNode) prometheus.PQLFilter {
		if node.IsChecked || len(node.Children) == 0 {
			return nil
		}

		var options []string
		for _, child := range node.Children {
			if child.IsChecked {
				options = append(options, child.Name)
			} else {
				if filter := dfs(child); filter != nil {
					filters = append(filters, filter)
				}
			}
		}

		if len(options) == 0 {
			return nil
		}

		switch node.Type {
		case datagroup.DATASOURCE_TYP_NAMESPACE:
			return prometheus.NewFilter().
				Equal("namespace", node.Name).
				Equal("cluster_id", node.ClusterID).
				RegexMatch("svc_name", prometheus.RegexMultipleValue(options...))
		case datagroup.DATASOURCE_TYP_CLUSTER:
			return prometheus.NewFilter().
				Equal("cluster_id", node.Name).
				RegexMatch("namespace", prometheus.RegexMultipleValue(options...))
		case datagroup.DATASOURCE_TYP_SYSTEM:
			return prometheus.NewFilter().
				RegexMatch("cluster_id", prometheus.RegexMultipleValue(options...))
		default:
			return nil
		}
	}

	if rootFilter := dfs(scopeNode); rootFilter != nil {
		filters = append(filters, rootFilter)
	}

	return prometheus.Or(filters...)
}
