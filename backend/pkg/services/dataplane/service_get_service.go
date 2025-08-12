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
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServices(ctx core.Context, req *request.QueryServicesRequest) *response.QueryServicesResponse {
	// Get the list of service
	services, _, err := s.chRepo.QueryServices(ctx, req.StartTime, req.EndTime, req.Cluster, 0, 0, nil)
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
	services, count, err := s.chRepo.QueryServices(ctx, req.StartTime, req.EndTime, req.Cluster, req.CurrentPage, req.PageSize, req.Services)
	if err != nil {
		return resp, err
	}

	filteredServices := make([]*model.Service, 0, len(services))
	serviceNames := make([]string, 0)

	for _, service := range services {
		found := false
		for _, datasource := range datasources {
			// source is empty means service is based on service rule
			if service.Source == datasource.Type || service.Source == "" {
				found = true
				break
			}
		}

		if found {
			filteredServices = append(filteredServices, service)
			serviceNames = append(serviceNames, service.Name)
		}
	}

	services = filteredServices
	filter := prometheus.NewFilter()
	filter.EqualIfNotEmpty("cluster_id", req.Cluster)
	filter.RegexMatch("service_name", prometheus.RegexMultipleValue(serviceNames...))
	instances, err := s.promRepo.GetDataplaneServiceInstances(ctx, req.StartTime, req.EndTime, filter.String())
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
