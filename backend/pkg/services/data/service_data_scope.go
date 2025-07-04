package data

import (
	"fmt"

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
