// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func GetGroupPQLFilter(ctx core.Context, dbRepo database.Repo, category string, groupID int64) (*prometheus.OrFilter, error) {
	groups, err := dbRepo.GetDataGroupByGroupIDOrUserID(ctx, groupID, ctx.UserID(), category)
	if err != nil {
		return nil, err
	}

	clusterNSList := make(map[string][]string, 0)
	clusterSvcList := make(map[string][]string, 0)
	for _, group := range groups {
		for _, ds := range group.DatasourceList {
			switch ds.Type {
			case model.DATASOURCE_TYP_SERVICE:
				if _, find := clusterSvcList[ds.ClusterID]; !find {
					clusterNSList[ds.ClusterID] = make([]string, 0)
				}
				svcList := clusterSvcList[ds.ClusterID]
				clusterSvcList[ds.ClusterID] = append(svcList, ds.Datasource)
			case model.DATASOURCE_TYP_NAMESPACE:
				if _, find := clusterNSList[ds.ClusterID]; !find {
					clusterSvcList[ds.ClusterID] = make([]string, 0)
				}
				nsList := clusterNSList[ds.ClusterID]
				clusterNSList[ds.ClusterID] = append(nsList, ds.Datasource)
			}
		}
	}

	var filters []prometheus.PQLFilter

	for clusterID, nsList := range clusterNSList {
		filter := prometheus.NewFilter().
			EqualIfNotEmpty("cluster_id", clusterID).
			RegexMatch("namespace", prometheus.RegexMultipleValue(nsList...))

		filters = append(filters, filter.(*prometheus.AndFilter))
	}

	for clusterID, svcList := range clusterSvcList {
		filter := prometheus.NewFilter().
			EqualIfNotEmpty("cluster_id", clusterID).
			RegexMatch(prometheus.ServiceNameKey, prometheus.RegexMultipleValue(svcList...))

		filters = append(filters, filter.(*prometheus.AndFilter))
	}

	return prometheus.Or(filters...), nil
}
