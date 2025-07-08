// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"fmt"
	"time"

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

	return &response.ListDataScopeFilterResponse{
		Scopes: scopes,
	}, nil
}

func (s *service) CleanExpiredDataScope(ctx core.Context, groupID int64, clean bool) (*response.CleanExpiredDataScopeResponse, error) {
	scopeIDs, err := common.ScanScope(ctx, s.promRepo, s.chRepo, s.dbRepo, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	options, err := s.dbRepo.GetScopeIDsOptionByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}
	permScopeIDs := common.DataGroupStorage.GetFullPermissionScopeList(options)

	for idx, id := range permScopeIDs {
		if id == "APO_ALL_DATA" {
			permScopeIDs = append(permScopeIDs[:idx], permScopeIDs[idx+1:]...)
			break
		}
	}

	userID := ctx.UserID()
	permGroupIDs, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	toClean := []string{}
	for _, id := range permScopeIDs {
		if _, ok := scopeIDs[id]; !ok {
			toClean = append(toClean, id)
		}
	}

	scopes, err := s.dbRepo.GetScopesByScopeIDs(ctx, toClean)
	if err != nil {
		return nil, err
	}

	permScopes, err := s.dbRepo.CheckScopesPermission(ctx, permGroupIDs, toClean)
	if err != nil {
		return nil, err
	}

	if clean {
		var deleteGroup2Scope = func(ctx core.Context) error {
			return s.dbRepo.DeleteGroup2ScopeByGroupIDScopeIDs(ctx, groupID, permScopes)
		}

		var deleteScope = func(ctx core.Context) error {
			return s.dbRepo.DeleteScopes(ctx, permScopes)
		}

		err = s.dbRepo.Transaction(ctx, deleteGroup2Scope, deleteScope)
		if err != nil {
			return nil, err
		}
	}

	var toCleanScopes []datagroup.DataScopeWithFullName
	var protectedScopes []datagroup.DataScopeWithFullName
	for _, scope := range scopes {
		if containsInStr(permScopes, scope.ScopeID) {
			toCleanScopes = append(toCleanScopes, datagroup.DataScopeWithFullName{
				DataScope: scope,
				FullName:  scope.FullName(),
			})
		} else {
			protectedScopes = append(protectedScopes, datagroup.DataScopeWithFullName{
				DataScope: scope,
				FullName:  scope.FullName(),
			})
		}
	}

	return &response.CleanExpiredDataScopeResponse{
		ToBeDeleted: toCleanScopes,
		Protected:   protectedScopes,
	}, nil
}
