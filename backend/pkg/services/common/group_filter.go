// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var (
	DataGroupStorage *DataGroupStore
	once             sync.Once
)

func InitDataGroupStorage(promRepo prometheus.Repo, chRepo clickhouse.Repo, dbRepo database.Repo) {
	once.Do(func() {
		DataGroupStorage = NewDatasourceStoreMap(promRepo, chRepo, dbRepo)
		DataGroupStorage.scanAndSave(core.EmptyCtx(), promRepo, chRepo, dbRepo, -48*time.Hour)
		DataGroupStorage.KeepWatchScope(core.EmptyCtx(), promRepo, chRepo, dbRepo, 10*time.Minute)
	})
}

func CutTopologyNodeInGroup(ctx core.Context, dbRepo database.Repo, groupID int64, topologyNode *model.TopologyNodes) (*model.TopologyNodes, error) {
	if groupID == 0 {
		return topologyNode, nil
	}

	selected, err := dbRepo.GetScopeIDsSelectedByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	svcList := DataGroupStorage.GetFullPermissionScopeList(selected)

	cutNode := &model.TopologyNodes{}
	for k, node := range topologyNode.Nodes {
		if node.Group != model.GROUP_SERVICE {
			continue
		}
		if containsInStr(svcList, node.Service) {
			cutNode.Nodes[k] = node
		}
	}
	return cutNode, nil
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
		default:
			return nil
		}
	}

	if rootFilter := dfs(scopeNode); rootFilter != nil {
		filters = append(filters, rootFilter)
	}

	return prometheus.Or(filters...)
}

func containsInStr(options []string, input string) bool {
	for _, v := range options {
		if v == input {
			return true
		}
	}
	return false
}
