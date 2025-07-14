// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceInstances(ctx core.Context, req *request.QueryServiceInstancesRequest) *response.QueryServiceInstancesResponse {
	filter := prometheus.NewFilter()
	filter.EqualIfNotEmpty("cluster_id", req.Cluster)
	filter.EqualIfNotEmpty("service_name", req.ServiceName)
	instances, err := s.promRepo.GetDataplaneServiceInstances(ctx, req.StartTime, req.EndTime, req.ServiceName, filter.String())
	if err != nil {
		return &response.QueryServiceInstancesResponse{
			Msg: "query service instances failed: " + err.Error(),
		}
	}
	return &response.QueryServiceInstancesResponse{
		Results: instances,
	}
}

func (s *service) getServiceInstancesByApo(ctx core.Context, req *request.QueryServiceInstancesRequest) *response.QueryServiceInstancesResponse {
	instanceList, err := s.promRepo.GetActiveInstanceList(ctx, req.StartTime, req.EndTime, []string{req.ServiceName})
	if err != nil {
		return &response.QueryServiceInstancesResponse{
			Msg: "query service instances failed: " + err.Error(),
		}
	}

	instances := make([]*model.ServiceInstance, 0)
	for _, instance := range instanceList.GetInstanceIdMap() {
		instances = append(instances, instance)
	}
	return &response.QueryServiceInstancesResponse{
		Results: instances,
	}
}
