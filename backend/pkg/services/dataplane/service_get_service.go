// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"fmt"
	"sort"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) GetServices(ctx core.Context, req *request.QueryServicesRequest) *response.QueryServicesResponse {
	var serviceNames []string
	if req.GroupID != 0 {
		selected, err := s.dbRepo.GetScopeIDsSelectedByGroupID(ctx, req.GroupID)
		if err != nil {
			return &response.QueryServicesResponse{
				Msg: "query data group scopes failed: " + err.Error(),
			}
		}
		for _, serviceName := range common.DataGroupStorage.GetFullPermissionSvcList(selected) {
			if serviceName != "" {
				serviceNames = append(serviceNames, serviceName)
			}
		}
		if len(serviceNames) == 0 {
			return &response.QueryServicesResponse{
				Results: []*model.Service{},
			}
		}
	}

	// Get the list of service
	services, _, err := s.chRepo.QueryServices(ctx, req.StartTime, req.EndTime, req.Cluster, 0, 0, serviceNames, nil)
	if err != nil {
		return &response.QueryServicesResponse{
			Msg: "query service failed: " + err.Error(),
		}
	}

	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	return &response.QueryServicesResponse{
		Results: services,
	}
}

func (s *service) queryServicesByApo(ctx core.Context, req *request.QueryServicesRequest) *response.QueryServicesResponse {
	services, err := s.promRepo.GetServiceList(ctx, req.StartTime, req.EndTime, nil)
	if err != nil {
		return &response.QueryServicesResponse{
			Msg: "query services failed: " + err.Error(),
		}
	}
	results := make([]*model.Service, 0)
	for _, service := range services {
		results = append(results, &model.Service{
			ClusterId: "",
			Source:    "apo",
			Id:        service,
			Name:      service,
		})
	}
	return &response.QueryServicesResponse{
		Results: results,
	}
}

func (s *service) GetServiceWithInstances(ctx core.Context, req *request.GetServiceWithInstancesRequest) (response.GetServiceWithInstancesResponse, error) {
	resp := response.GetServiceWithInstancesResponse{
		Services: make([]response.ServiceWithInstances, 0),
	}
	datasources, err := s.dbRepo.GetDatapalenDsCount(ctx)
	if err != nil {
		return resp, err
	}

	if len(datasources) == 0 {
		return resp, nil
	}

	// Get the list of service
	sources := make([]string, 0)
	for _, datasource := range datasources {
		sources = append(sources, datasource.Type)
	}

	services, count, err := s.chRepo.QueryServices(ctx, req.StartTime, req.EndTime, req.Cluster, req.CurrentPage, req.PageSize, req.Services, sources)
	if err != nil {
		return resp, err
	}

	serviceNames := make([]string, 0, len(services))

	for _, service := range services {
		serviceNames = append(serviceNames, service.Name)
	}
	instances, err := s.chRepo.GetDataplaneServiceInstances(ctx, req.StartTime, req.EndTime, req.Cluster, serviceNames)
	if err != nil {
		return resp, err
	}

	instancesLabels, err := s.chRepo.GetInstancesLabels(ctx, req.StartTime, req.EndTime, instances)
	if err != nil {
		return resp, err
	}

	servicesWithInstances := make([]response.ServiceWithInstances, 0, len(services))

	instancesByService := make(map[string]map[string]model.ServiceInstance)
	for _, instance := range instances {
		instanceName := instance.GetInstanceId()

		if instancesByService[instance.ServiceName] == nil {
			instancesByService[instance.ServiceName] = make(map[string]model.ServiceInstance)
		}

		instancesByService[instance.ServiceName][instanceName] = *instance
	}

	for _, service := range services {
		serviceWithInstances := response.ServiceWithInstances{
			Service:   *service,
			Instances: make([]response.InstanceWithSource, 0),
		}

		if serviceInstancesMap, exists := instancesByService[service.Name]; exists {
			for instanceName, instance := range serviceInstancesMap {
				labelKey := fmt.Sprintf("%s-%s-%s-%s-%d",
					instance.ServiceName,
					instance.PodName,
					instance.NodeName,
					instance.ContainerId,
					instance.Pid,
				)

				instanceWithSource := response.InstanceWithSource{
					ServiceInstance: instance,
					InstanceName:    instanceName,
					InstanceSource:  determineInstanceSource(instancesLabels[labelKey]),
				}

				serviceWithInstances.Instances = append(serviceWithInstances.Instances, instanceWithSource)
			}
		}

		servicesWithInstances = append(servicesWithInstances, serviceWithInstances)
	}

	sort.Slice(servicesWithInstances, func(i, j int) bool {
		return servicesWithInstances[i].Service.Name < servicesWithInstances[j].Service.Name
	})
	resp.Services = servicesWithInstances
	resp.PageSize = req.PageSize
	resp.CurrentPage = req.CurrentPage
	resp.Total = int64(count)

	return resp, nil
}

func determineInstanceSource(labels map[string]string) (instanceSource response.InstanceSource) {
	ruleId, exit := labels["rule_id"]
	if exit {
		instanceSource.Type = "rule"
		instanceSource.Name = ruleId
		return
	}
	source, exist := labels["source"]
	if exist {
		instanceSource.Type = "source"
		instanceSource.Name = source
		return
	}

	return
}
