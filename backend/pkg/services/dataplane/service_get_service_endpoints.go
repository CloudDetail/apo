// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceEndpoints(ctx core.Context, req *request.QueryServiceEndpointsRequest) *response.QueryServiceEndpointsResponse {
	// Get the list of service Endpoint
	filter := prometheus.EqualFilter(prometheus.ServiceNameKey, req.ServiceName)

	endpoints, err := s.promRepo.GetServiceEndPointListByPQLFilter(ctx, req.StartTime, req.EndTime, filter)
	if err != nil {
		return &response.QueryServiceEndpointsResponse{
			Msg: "query service endpoints failed: " + err.Error(),
		}
	}
	if len(endpoints) == 0 {
		return s.queryServiceEndpointsByApi(ctx, req)
	}

	return &response.QueryServiceEndpointsResponse{
		Results: endpoints,
	}
}

func (s *service) queryServiceEndpointsByApi(ctx core.Context, req *request.QueryServiceEndpointsRequest) *response.QueryServiceEndpointsResponse {
	return &response.QueryServiceEndpointsResponse{
		Msg: "Data not found",
	}
}
