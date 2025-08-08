// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServices(ctx core.Context, req *request.QueryServicesRequest) *response.QueryServicesResponse {
	// Get the list of service
	services, count, err := s.chRepo.QueryServices(ctx, req.StartTime, req.EndTime, req.Cluster, req.CurrentPage, req.PageSize)
	if err != nil {
		return &response.QueryServicesResponse{
			Msg: "query service failed: " + err.Error(),
		}
	}
	return &response.QueryServicesResponse{
		Results: services,
		Pagination: model.Pagination{
			PageSize:    req.PageSize,
			CurrentPage: req.CurrentPage,
			Total:       int64(count),
		},
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
	now := time.Now()
	startOfDay := now.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)
	startOfDayTimestamp := startOfDay.UnixMicro()
	endOfDayTimestamp := endOfDay.UnixMicro()

	resp := response.GetServiceWithInstancesResponse{
		Services: make([]response.ServiceWithInstances, 0),
	}
	services, count, err := s.chRepo.QueryServices(ctx, req.StartTime, req.EndTime, req.Cluster, req.CurrentPage, req.PageSize)
	if err != nil {
		return resp, err
	}

	serviceNames := make([]string, 0, len(services))
	serviceMap := make(map[string]*model.Service)
	for _, service := range services {
		serviceNames = append(serviceNames, service.Name)
		serviceMap[service.Name] = service
	}

	filter := prometheus.NewFilter()
	filter.EqualIfNotEmpty("cluster_id", req.Cluster)
	filter.RegexMatch("service_name", prometheus.RegexMultipleValue(serviceNames...))
	instances, err := s.promRepo.GetDataplaneServiceInstances(ctx, req.StartTime, req.EndTime, filter.String())
	if err != nil {
		return resp, err
	}

	instancesLabels, err := s.chRepo.GetInstancesLabels(ctx, startOfDayTimestamp, endOfDayTimestamp, instances)
	if err != nil {
		return resp, err
	}

	servicesWithInstances := make([]response.ServiceWithInstances, 0, len(services))

	// 按服务名分组instances
	instancesByService := make(map[string][]model.ServiceInstance)
	for _, instance := range instances {
		instancesByService[instance.ServiceName] = append(instancesByService[instance.ServiceName], *instance)
	}

	for _, service := range services {
		serviceWithInstances := response.ServiceWithInstances{
			Service:   *service,
			Instances: make([]response.InstanceWithSource, 0),
		}

		// 获取该服务的所有实例
		if serviceInstances, exists := instancesByService[service.Name]; exists {
			for _, instance := range serviceInstances {
				// 生成instance labels的key: serviceName-podName-nodeName-containerId
				labelKey := fmt.Sprintf("%s-%s-%s-%s-%d",
					instance.ServiceName,
					instance.PodName,
					instance.NodeName,
					instance.ContainerId,
					instance.Pid,
				)

				instanceWithSource := response.InstanceWithSource{
					ServiceInstance: instance,
					InstanceName:    instance.GetInstanceId(),
					InstanceSource:  determineInstanceSource(instance, instancesLabels[labelKey]),
				}

				serviceWithInstances.Instances = append(serviceWithInstances.Instances, instanceWithSource)
			}
		}

		servicesWithInstances = append(servicesWithInstances, serviceWithInstances)
	}

	resp.Services = servicesWithInstances
	resp.PageSize = req.PageSize
	resp.CurrentPage = req.CurrentPage
	resp.Total = int64(count)

	return resp, nil
}

func determineInstanceSource(instance model.ServiceInstance, labels map[string]string) (instanceSource response.InstanceSource) {
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
