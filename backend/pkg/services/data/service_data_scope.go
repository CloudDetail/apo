package data

import (
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
		scopes = s.DataGroupStore.CloneScopeWithPermission(selected, nil)
	} else {
		scopes = s.DataGroupStore.CloneScopeWithPermission(options, selected)
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

	scopes, leafs := s.DataGroupStore.CloneWithCategory(scopeIDs, req.Category)
	filter := common.ConvertScopeNodeToPQLFilter(scopes)

	switch req.Category {
	case datagroup.DATASOURCE_CATEGORY_APM:
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
				node.Children = append(node.Children, &datagroup.DataScopeTreeNode{
					DataScope: datagroup.DataScope{
						Name: metric.Metric.ContentKey,
						Type: datagroup.DATASOURCE_TYP_ENDPOINT,
					},
				})
			}
		}

	case datagroup.DATASOURCE_CATEGORY_LOG:
		series, err := s.promRepo.QueryMetricsWithPQLFilter(
			ctx, prometheus.PQLMetricSeries(prometheus.LOG_EXCEPTION_COUNT, prometheus.LOG_LEVEL_COUNT),
			req.StartTime, req.EndTime,
			"cluster_id,namespace,pod", filter,
		)

		if err != nil {
			return nil, err
		}

		for _, metric := range series {
			label := datagroup.ScopeLabels{
				ClusterID: metric.Metric.ClusterID,
				Namespace: metric.Metric.Namespace,
			}

			node := leafs[label]
			if node != nil {
				node.Children = append(node.Children, &datagroup.DataScopeTreeNode{
					DataScope: datagroup.DataScope{
						Name: metric.Metric.POD,
						Type: datagroup.DATASOURCE_TYP_POD,
					},
				})
			}
		}
	}

	return &response.ListDataScopeFilterResponse{
		Scopes: scopes,
	}, nil
}
