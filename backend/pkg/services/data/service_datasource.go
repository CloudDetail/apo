package data

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
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

//func (s *service) GetGroupDatasource(req *request.GetGroupDatasourceRequest) (response.GetGroupDatasourceResponse, error) {
//	var resp response.GetGroupDatasourceResponse
//	filter := model.DataGroupFilter{
//		ID: req.GroupID,
//	}
//
//	dataGroup, _, err := s.dbRepo.GetDataGroup(filter)
//	if err != nil {
//		return resp, err
//	}
//
//	if len(dataGroup) == 0 {
//		return resp, model.NewErrWithMessage(errors.New("data group does not exits"), code.DataGroupNotExistError)
//	}
//
//	return resp, nil
//}
