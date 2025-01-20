package data

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"time"
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
		ds := model.Datasource{
			Datasource: namespace,
			Type:       model.DATASOURCE_TYP_NAMESPACE,
			Category:   model.DATASOURCE_CATEGORY_NORMAL,
			Nested:     services,
		}
		namespaceList = append(namespaceList, ds)
	}

	resp.NamespaceList = namespaceList
	resp.ServiceList = serviceList
	return resp, nil
}

func (s *service) GetGroupDatasource(req *request.GetGroupDatasourceRequest, userID int64) (response.GetGroupDatasourceResponse, error) {
	var (
		services     []string
		groups       []database.DataGroup
		err          error
		namespaceMap = map[string][]string{}
		resp         = response.GetGroupDatasourceResponse{}
	)
	if req.GroupID != 0 {
		groups, err = s.getDataGroup(req.GroupID)
	} else {
		groups, err = s.getUserDataGroup(userID, req.Category)
	}

	if err != nil {
		return resp, err
	}

	for _, group := range groups {
		for _, ds := range group.DatasourceList {
			nested, err := s.getNested(ds.Datasource, ds.Type)
			if err != nil {
				return resp, err
			}

			if ds.Type == model.DATASOURCE_TYP_NAMESPACE {
				namespaceMap[ds.Datasource] = nested
			} else if ds.Type == model.DATASOURCE_TYP_SERVICE {
				for _, namespace := range nested {
					namespaceMap[namespace] = append(namespaceMap[namespace], ds.Datasource)
					services = append(services, ds.Datasource)
				}
			}
		}
	}

	resp.NamespaceMap = namespaceMap
	resp.ServiceList = services
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

func (s *service) getDataGroup(groupID int64) ([]database.DataGroup, error) {
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

	return dataGroups, nil
}
