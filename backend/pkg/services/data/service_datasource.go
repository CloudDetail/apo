// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"sort"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

var subTime = -time.Hour * 24 * 15

func (s *service) GetDataSource(ctx_core core.Context) (resp response.GetDatasourceResponse, err error) {
	var (
		endTime        = time.Now()
		startTime      = endTime.Add(subTime)
		endTimeMicro   = endTime.UnixMicro()
		startTimeMicro = startTime.UnixMicro()
	)

	servicesMap, err := s.promRepo.GetServiceWithNamespace(startTimeMicro, endTimeMicro, nil)
	if err != nil {
		return
	}

	namespaceMap, err := s.promRepo.GetNamespaceWithService(startTimeMicro, endTimeMicro)
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

func (s *service) GetGroupDatasource(ctx_core core.Context, req *request.GetGroupDatasourceRequest, userID int64) (response.GetGroupDatasourceResponse, error) {
	var (
		groups       []database.DataGroup
		err          error
		namespaceMap = map[string][]string{}
		serviceMap   = map[string][]string{}
		filterMap    = map[string]struct{}{}
		resp         = response.GetGroupDatasourceResponse{}
		endTime      = time.Now()
		startTime    = endTime.Add(subTime)
	)
	if req.GroupID != 0 {
		groups, err = s.getDataGroup(ctx_core, req.GroupID, req.Category)
	} else {
		groups, err = s.getUserDataGroup(ctx_core, userID, req.Category)
	}

	if len(groups) == 0 {
		defaultGroup, err := s.getDefaultDataGroup(ctx_core, req.Category)
		if err != nil {
			return resp, err
		}

		groups = append(groups, defaultGroup)
	}

	if err != nil {
		return resp, err
	}

	for _, group := range groups {
		for _, ds := range group.DatasourceList {
			if ds.Type == model.DATASOURCE_TYP_NAMESPACE {
				if len(ds.Datasource) == 0 {
					continue
				}
				namespaceMap[ds.Datasource] = []string{}
			} else if ds.Type == model.DATASOURCE_TYP_SERVICE {
				serviceMap[ds.Datasource] = []string{}
			}
		}
	}

	for namespace := range namespaceMap {
		nested, err := s.getNested(namespace, model.DATASOURCE_TYP_NAMESPACE)
		if err != nil {
			return resp, err
		}
		namespaceMap[namespace] = nested

		for _, srv := range nested {
			filterMap[namespace+srv] = struct{}{}
			serviceMap[srv] = []string{}
		}
	}

	for service := range serviceMap {
		nested, err := s.getNested(service, model.DATASOURCE_TYP_SERVICE)
		if err != nil {
			return resp, err
		}
		for _, namespace := range nested {
			if _, ok := filterMap[namespace+service]; ok || len(namespace) == 0 {
				continue
			}
			namespaceMap[namespace] = append(namespaceMap[namespace], service)
			filterMap[namespace+service] = struct{}{}
		}
		endpoints, err := s.promRepo.GetServiceEndPointList(startTime.UnixMicro(), endTime.UnixMicro(), service)
		if err != nil {
			return resp, err
		}

		serviceMap[service] = endpoints
	}

	resp.NamespaceMap = namespaceMap
	resp.ServiceMap = serviceMap
	return resp, nil
}

func (s *service) getNested(datasource string, typ string) ([]string, error) {
	var (
		endTime   = time.Now()
		startTime = endTime.Add(-24 * time.Hour)
		nested    []string
		err       error
	)

	if typ == model.DATASOURCE_TYP_NAMESPACE {
		nested, err = s.promRepo.GetServiceList(startTime.UnixMicro(), endTime.UnixMicro(), []string{datasource})
	} else if typ == model.DATASOURCE_TYP_SERVICE {
		nested, err = s.promRepo.GetServiceNamespace(startTime.UnixMicro(), endTime.UnixMicro(), datasource)
	}

	return nested, err
}

func (s *service) getDataGroup(ctx_core core.Context, groupID int64, category string) ([]database.DataGroup, error) {
	filter := model.DataGroupFilter{
		ID: groupID,
	}

	dataGroups, _, err := s.dbRepo.GetDataGroup(ctx_core, filter)
	if err != nil {
		return dataGroups, err
	}

	if len(dataGroups) == 0 {
		return nil, core.Error(code.DataGroupNotExistError, "data group does not exits")
	}

	for i, group := range dataGroups {
		filteredDatasource := make([]database.DatasourceGroup, 0, len(group.DatasourceList))
		for _, ds := range group.DatasourceList {
			if len(category) == 0 || category == ds.Category {
				filteredDatasource = append(filteredDatasource, ds)
			}
		}
		dataGroups[i].DatasourceList = filteredDatasource
	}

	return dataGroups, nil
}
