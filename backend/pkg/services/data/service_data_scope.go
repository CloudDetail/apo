// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"fmt"
	"slices"
	"time"

	"github.com/CloudDetail/apo/backend/config"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) ListDataScopeByGroupID(ctx core.Context, req *request.DGScopeListRequest) (*response.ListDataScopesResponse, error) {
	options, err := s.dbRepo.GetScopeIDsOptionByGroupID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	selected, err := s.dbRepo.GetScopeIDsSelectedByGroupID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	var scopes *datagroup.DataScopeTreeNode
	if req.SkipNotChecked {
		scopes = common.DataGroupStorage.CloneScopeWithPermission(selected, nil)
	} else {
		scopes = common.DataGroupStorage.CloneScopeWithPermission(options, selected)
	}

	clusterNameMap, err := s.dbRepo.ListClusterName(ctx)
	if err == nil && clusterNameMap != nil {
		scopes.FillWithClusterName(clusterNameMap)
	}

	return &response.ListDataScopesResponse{
		Scopes:      scopes,
		DataSources: selected,
	}, nil
}

func (s *service) GetFilterByGroupID(ctx core.Context, req *request.DGFilterRequest) (*response.ListDataScopeFilterResponse, error) {
	scopeIDs, err := s.dbRepo.GetScopeIDsByGroupIDAndCat(ctx, req.GroupID, req.Category)
	if err != nil {
		return nil, err
	}

	if len(scopeIDs) == 0 {
		return &response.ListDataScopeFilterResponse{
			Scopes: &datagroup.DataScopeTreeNode{},
		}, nil
	}

	scopes, leafs := common.DataGroupStorage.CloneWithCategory(scopeIDs, req.Category)

	filter := common.ConvertScopeNodeToPQLFilter(scopes)

	switch req.Extra {
	case "endpoint":
		series, err := s.promRepo.QueryMetricsWithPQLFilter(
			ctx, prometheus.PQLMetricSeries(prometheus.SPAN_TRACE_COUNT),
			req.StartTime, req.EndTime,
			"cluster_id,namespace,svc_name,content_key", filter,
		)

		if err != nil {
			return nil, err
		}

		for _, metric := range series {
			label := datagroup.ScopeLabels{
				ClusterID: metric.Metric.ClusterID,
				Namespace: metric.Metric.Namespace,
				Service:   metric.Metric.SvcName,
			}

			node := leafs[label]
			if node != nil {
				node.ExtraChildren = append(node.ExtraChildren, &datagroup.ExtraChild{
					ID:       fmt.Sprintf("%s#%s", node.ScopeID, metric.Metric.ContentKey),
					Name:     metric.Metric.ContentKey,
					Type:     req.Extra,
					Endpoint: metric.Metric.ContentKey,
					Service:  metric.Metric.SvcName,
				})
			}
		}

	case "instance":
		series, err := s.promRepo.QueryMetricsWithPQLFilter(
			ctx,
			prometheus.LogErrorCountSeriesCombineSvcInfoWithPQLFilter,
			req.StartTime, req.EndTime,
			"cluster_id,namespace,svc_name,pod,node_name,pid,container_id", filter,
		)

		if err != nil {
			return nil, err
		}

		for _, metric := range series {
			label := datagroup.ScopeLabels{
				ClusterID: metric.Metric.ClusterID,
				Namespace: metric.Metric.Namespace,
				Service:   metric.Metric.SvcName,
			}

			node := leafs[label]
			if node != nil {
				var extraName string
				if len(metric.Metric.POD) > 0 {
					extraName = metric.Metric.POD
				} else {
					extraName = fmt.Sprintf("%s#%s", metric.Metric.NodeName, metric.Metric.ContainerID)
				}

				node.ExtraChildren = append(node.ExtraChildren, &datagroup.ExtraChild{
					ID:          extraName,
					Name:        extraName,
					Type:        req.Extra,
					ContainerID: metric.Metric.ContainerID,
					POD:         metric.Metric.POD,
					Node:        metric.Metric.NodeName,
					Pid:         metric.Metric.PID,
					Service:     metric.Metric.SvcName,
				})
			}
		}
	}

	clusterNameMap, err := s.dbRepo.ListClusterName(ctx)
	if err != nil {
		return nil, err
	}
	scopes.FillWithClusterName(clusterNameMap)

	return &response.ListDataScopeFilterResponse{
		Scopes: scopes,
	}, nil
}

func (s *service) CleanExpiredDataScope(ctx core.Context, groupID int64, clean bool) (*response.CleanExpiredDataScopeResponse, error) {
	cfg := config.Get().DataGroup
	if cfg.InitLookBackDays <= 0 {
		cfg.InitLookBackDays = 3
	}

	realTimeScopeIDs, err := common.ScanScope(ctx, s.promRepo, s.chRepo, s.dbRepo, time.Duration(cfg.InitLookBackDays)*24*time.Hour)
	if err != nil {
		return nil, err
	}

	toBeCheckScopeIDs, err := s.dbRepo.GetScopeIDsOptionByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	toBeCheckScopeIDs = common.DataGroupStorage.GetFullPermissionScopeList(toBeCheckScopeIDs)

	// SKIP APO_ALL_DATA SCOPE
	for idx, id := range toBeCheckScopeIDs {
		if id == "APO_ALL_DATA" {
			toBeCheckScopeIDs = append(toBeCheckScopeIDs[:idx], toBeCheckScopeIDs[idx+1:]...)
			break
		}
	}

	needCleanScopeIDs := []string{}
	for _, id := range toBeCheckScopeIDs {
		if _, ok := realTimeScopeIDs[id]; !ok {
			needCleanScopeIDs = append(needCleanScopeIDs, id)
		}
	}

	needCleanScopes, err := s.dbRepo.GetScopesByScopeIDs(ctx, needCleanScopeIDs)
	if err != nil {
		return nil, err
	}

	userID := ctx.UserID()
	permGroupIDs, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}
	permGroups := common.DataGroupStorage.GetFullPermissionGroup(permGroupIDs)
	for _, group := range permGroups {
		permGroupIDs = append(permGroupIDs, group.GroupID)
	}

	// Check Group Permission
	noPermScopes, err := s.dbRepo.PeekScopeIDWithoutPerm(ctx, permGroupIDs, needCleanScopeIDs)
	if err != nil {
		return nil, err
	}

	var toCleanScopes, protectedScopes []datagroup.DataScopeWithFullName
	var toCleanScopeIDs []string
	for _, scope := range needCleanScopes {
		if slices.Contains(noPermScopes, scope.ScopeID) {
			protectedScopes = append(protectedScopes, datagroup.DataScopeWithFullName{
				DataScope: scope,
				FullName:  scope.FullName(),
			})
		} else {
			toCleanScopeIDs = append(toCleanScopeIDs, scope.ScopeID)
			toCleanScopes = append(toCleanScopes, datagroup.DataScopeWithFullName{
				DataScope: scope,
				FullName:  scope.FullName(),
			})
		}
	}

	if clean {
		var deleteGroup2Scope = func(ctx core.Context) error {
			return s.dbRepo.DeleteGroup2ScopeByGroupIDScopeIDs(ctx, groupID, toCleanScopeIDs)
		}

		var deleteScope = func(ctx core.Context) error {
			return s.dbRepo.DeleteScopes(ctx, toCleanScopeIDs)
		}

		err = s.dbRepo.Transaction(ctx, deleteGroup2Scope, deleteScope)
		if err != nil {
			return nil, err
		}

		// Refresh ScopeTree
		_ = common.DataGroupStorage.Refresh(ctx, s.promRepo, s.chRepo, s.dbRepo, 10*time.Minute)
		newScopeTree, err := s.dbRepo.LoadScopes(ctx)
		if err == nil {
			common.DataGroupStorage.DataScopeTree = newScopeTree
		}
	}

	return &response.CleanExpiredDataScopeResponse{
		ToBeDeleted: toCleanScopes,
		Protected:   protectedScopes,
	}, nil
}
