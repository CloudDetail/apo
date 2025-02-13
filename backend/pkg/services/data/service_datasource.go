// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (s *service) GetDataSource() (resp response.GetDatasourceResponse, err error) {
	var (
		endTime        = time.Now()
		startTime      = endTime.Add(-24 * time.Hour)
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
			return namespaces[i] < namespaces[j]
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
		sort.Slice(services, func(i, j int) bool {
			return services[i] < services[j]
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
		return namespaceList[i].Datasource < namespaceList[j].Datasource
	})

	resp.NamespaceList = namespaceList
	resp.ServiceList = serviceList
	return resp, nil
}

func (s *service) GetGroupDatasource(req *request.GetGroupDatasourceRequest, userID int64) (response.GetGroupDatasourceResponse, error) {
	var (
		groups       []database.DataGroup
		err          error
		namespaceMap = map[string][]string{}
		serviceMap   = map[string][]string{}
		resp         = response.GetGroupDatasourceResponse{}
		endTime      = time.Now()
		startTime    = endTime.Add(-time.Hour * 24)
	)
	if req.GroupID != 0 {
		groups, err = s.getDataGroup(req.GroupID, req.Category)
	} else {
		groups, err = s.getUserDataGroup(userID, req.Category)
	}

	if len(groups) == 0 {
		defaultGroup, err := s.getDefaultDataGroup(req.Category)
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
			nested := make([]string, 0)
			if ds.Category == model.DATASOURCE_CATEGORY_APM {
				nested, err = s.getNested(ds.Datasource, ds.Type)
				if err != nil {
					return resp, err
				}
			}

			if ds.Type == model.DATASOURCE_TYP_NAMESPACE {
				namespaceMap[ds.Datasource] = nested
				for _, srv := range nested {
					endpoints, err := s.promRepo.GetServiceEndPointList(startTime.UnixMicro(), endTime.UnixMicro(), srv)
					if err != nil {
						return response.GetGroupDatasourceResponse{}, err
					}

					serviceMap[srv] = endpoints
				}
			} else if ds.Type == model.DATASOURCE_TYP_SERVICE {
				for _, namespace := range nested {
					namespaceMap[namespace] = append(namespaceMap[namespace], ds.Datasource)
				}

				endpoints, err := s.promRepo.GetServiceEndPointList(startTime.UnixMicro(), endTime.UnixMicro(), ds.Datasource)
				if err != nil {
					return response.GetGroupDatasourceResponse{}, err
				}
				serviceMap[ds.Datasource] = endpoints
			}
		}
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

func (s *service) getDataGroup(groupID int64, category string) ([]database.DataGroup, error) {
	filter := model.DataGroupFilter{
		ID: groupID,
	}

	dataGroups, _, err := s.dbRepo.GetDataGroup(filter)
	if err != nil {
		return dataGroups, err
	}

	if len(dataGroups) == 0 {
		return nil, model.NewErrWithMessage(errors.New("data group does not exits"), code.DataGroupNotExistError)
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
