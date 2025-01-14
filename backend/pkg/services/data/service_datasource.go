package data

import (
	"errors"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"time"
)

func (s *service) GetDataSource() (resp response.GetDatasourceResponse, err error) {
	namespaces, err := s.k8sRepo.GetNamespaceList()
	if err != nil {
		return nil, err
	}

	endTime := time.Now()
	startTime := endTime.Add(-24 * time.Hour)
	endTimeMicro := endTime.UnixMicro()
	startTimeMicro := startTime.UnixMicro()
	services, err := s.promRepo.GetServiceList(startTimeMicro, endTimeMicro, nil)
	if err != nil {
		return nil, err
	}

	apmNamespace, err := s.promRepo.GetNamespaceList(startTimeMicro, endTimeMicro)
	if err != nil {
		return nil, err
	}

	apmNamespaceMap := make(map[string]struct{})
	for _, apmNs := range apmNamespace {
		apmNamespaceMap[apmNs] = struct{}{}
	}

	datasource := make([]model.Datasource, 0, len(namespaces.Items)+len(services))
	for _, namespace := range namespaces.Items {
		category := model.DATASOURCE_CATEGORY_NORMAL
		if _, isAPM := apmNamespaceMap[namespace.Name]; isAPM {
			category = model.DATASOURCE_CATEGORY_APM
		}
		datasource = append(datasource, model.Datasource{
			Datasource: namespace.Name,
			Type:       model.DATASOURCE_TYP_NAMESPACE,
			Category:   category,
		})
	}

	for _, service := range services {
		datasource = append(datasource, model.Datasource{
			Datasource: service,
			Type:       model.DATASOURCE_TYP_SERVICE,
			Category:   model.DATASOURCE_CATEGORY_APM,
		})
	}

	return datasource, nil
}

func (s *service) GetGroupDatasource(req *request.GetGroupDatasourceRequest) (response.GetGroupDatasourceResponse, error) {
	var resp response.GetGroupDatasourceResponse
	filter := model.DataGroupFilter{
		ID: req.GroupID,
	}

	dataGroup, _, err := s.dbRepo.GetDataGroup(filter)
	if err != nil {
		return resp, err
	}

	if len(dataGroup) == 0 {
		return resp, model.NewErrWithMessage(errors.New("data group does not exits"), code.DataGroupNotExistError)
	}

	resp = response.GetGroupDatasourceResponse(dataGroup[0])
	return resp, nil
}
