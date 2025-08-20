// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceInstances(ctx core.Context, req *request.QueryServiceInstancesRequest) *response.QueryServiceInstancesResponse {
	instances, err := s.chRepo.GetDataplaneServiceInstances(ctx, req.StartTime, req.EndTime, req.Cluster, []string{req.ServiceName})
	if err != nil {
		return &response.QueryServiceInstancesResponse{
			Msg: "query service instances failed: " + err.Error(),
		}
	}
	if len(instances) > 0 {
		return &response.QueryServiceInstancesResponse{
			Results: instances,
		}
	}
	return s.getServiceInstancesByApo(ctx, req)
}

func (s *service) getServiceInstancesByApo(ctx core.Context, req *request.QueryServiceInstancesRequest) *response.QueryServiceInstancesResponse {
	instanceList, err := s.promRepo.GetActiveInstanceList(ctx, req.StartTime, req.EndTime, req.Cluster, []string{req.ServiceName})
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
