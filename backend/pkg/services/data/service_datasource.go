// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"sort"
	"strings"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

var subTime = -time.Hour * 24 * 15

func (s *service) GetDataSource(ctx core.Context) (resp response.GetDatasourceResponse, err error) {
	var (
		endTime        = time.Now()
		startTime      = endTime.Add(subTime)
		endTimeMicro   = endTime.UnixMicro()
		startTimeMicro = startTime.UnixMicro()
	)

	servicesMap, err := s.promRepo.GetServiceWithNamespace(ctx, startTimeMicro, endTimeMicro, nil)
	if err != nil {
		return
	}

	namespaceMap, err := s.promRepo.GetNamespaceWithService(ctx, startTimeMicro, endTimeMicro)
	if err != nil {
		return
	}

	allNamespaces, err := s.k8sRepo.GetNamespaceList()
	if err != nil {
		return
	}

	serviceList := make([]model.Datasource, 0, len(servicesMap))
	for service, namespaces := range servicesMap {
		sort.Slice(namespaces, func(i, j int) bool {
			return strings.Compare(namespaces[i], namespaces[j]) < 0
		})
		ds := model.Datasource{
			Datasource: service,
			Type:       model.DATASOURCE_TYP_SERVICE,
			Category:   model.DATASOURCE_CATEGORY_APM,
			Nested:     namespaces,
		}
		serviceList = append(serviceList, ds)
	}

	namespaceList := make([]model.Datasource, 0)
	for _, namespace := range allNamespaces.Items {
		if _, ok := namespaceMap[namespace.Name]; !ok {
			ds := model.Datasource{
				Datasource: namespace.Name,
				Type:       model.DATASOURCE_TYP_NAMESPACE,
				Category:   model.DATASOURCE_CATEGORY_NORMAL,
			}
			namespaceList = append(namespaceList, ds)
		}
	}

	for namespace, services := range namespaceMap {
		if len(namespace) == 0 {
			continue
		}
		sort.Slice(services, func(i, j int) bool {
			return strings.Compare(services[i], services[j]) < 0
		})
		ds := model.Datasource{
			Datasource: namespace,
			Type:       model.DATASOURCE_TYP_NAMESPACE,
			Category:   model.DATASOURCE_CATEGORY_APM,
			Nested:     services,
		}
		namespaceList = append(namespaceList, ds)
	}

	sort.Slice(serviceList, func(i, j int) bool {
		return strings.Compare(serviceList[i].Datasource, serviceList[j].Datasource) < 0
	})

	sort.Slice(namespaceList, func(i, j int) bool {
		return strings.Compare(namespaceList[i].Datasource, namespaceList[j].Datasource) < 0
	})

	resp.NamespaceList = namespaceList
	resp.ServiceList = serviceList
	return resp, nil
}

func (s *service) GetGroupDatasource(ctx core.Context, req *request.GetGroupDatasourceRequest) (response.GetGroupDatasourceResponse, error) {
	filter, err := common.GetPQLFilterByGroupID(ctx, s.dbRepo, req.Category, req.GroupID)
	if err != nil {
		return response.GetGroupDatasourceResponse{}, err
	}

	endTime := time.Now()
	startTime := endTime.Add(subTime)

	labels, err := s.promRepo.QueryMetricsWithPQLFilter(
		ctx,
		prometheus.PQLMetricSeries(prometheus.SPAN_TRACE_COUNT),
		startTime.UnixMicro(),
		endTime.UnixMicro(),
		"cluster_id,namespace,svc_name,content_key",
		filter,
	)

	if err != nil {
		return response.GetGroupDatasourceResponse{}, err
	}

	return response.GetGroupDatasourceResponse{
		GroupDatasource:         groupScopeMap(labels),
		ClusterScopedDatasource: clusterScopeMap(labels, nil), // TODO pass clusterName map
	}, nil
}

func clusterScopeMap(labelsList []prometheus.MetricResult, clusterNames map[string]string) []response.ClusterScopedDatasource {
	clusterMap := make(map[string]*response.ClusterScopedDatasource)
	for _, l := range labelsList {
		clusterID := l.Metric.ClusterID
		clusterName := clusterID
		if clusterNames != nil {
			if name, find := clusterNames[clusterID]; find {
				clusterName = name
			}
		}
		namespace := l.Metric.Namespace
		svc := l.Metric.SvcName
		endpoint := l.Metric.ContentKey

		if _, ok := clusterMap[clusterID]; !ok {
			clusterMap[clusterID] = &response.ClusterScopedDatasource{
				ClusterID:   clusterID,
				ClusterName: clusterName,
				GroupDatasource: response.GroupDatasource{
					NamespaceMap: make(map[string][]string),
					ServiceMap:   make(map[string][]string),
				},
			}
		}

		ds := clusterMap[clusterID]

		if !contains(ds.NamespaceMap[namespace], svc) {
			ds.NamespaceMap[namespace] = append(ds.NamespaceMap[namespace], svc)
		}

		if !contains(ds.ServiceMap[svc], endpoint) {
			ds.ServiceMap[svc] = append(ds.ServiceMap[svc], endpoint)
		}
	}

	var result []response.ClusterScopedDatasource
	for _, v := range clusterMap {
		result = append(result, *v)
	}
	return result
}

func groupScopeMap(labelsList []prometheus.MetricResult) response.GroupDatasource {
	ds := response.GroupDatasource{
		NamespaceMap: make(map[string][]string),
		ServiceMap:   make(map[string][]string),
	}

	for _, l := range labelsList {
		namespace := l.Metric.Namespace
		svc := l.Metric.SvcName
		endpoint := l.Metric.ContentKey

		if !contains(ds.NamespaceMap[namespace], svc) {
			ds.NamespaceMap[namespace] = append(ds.NamespaceMap[namespace], svc)
		}

		if !contains(ds.ServiceMap[svc], endpoint) {
			ds.ServiceMap[svc] = append(ds.ServiceMap[svc], endpoint)
		}
	}

	return ds
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
