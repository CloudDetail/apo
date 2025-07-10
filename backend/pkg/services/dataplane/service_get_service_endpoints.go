// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceEndpoints(ctx core.Context, req *request.QueryServiceEndpointsRequest) *response.QueryServiceEndpointsResponse {
	endpoints, err := s.chRepo.QueryServiceEndpoints(ctx, req.StartTime, req.EndTime, req.Cluster, req.ServiceName)
	if err != nil {
		return &response.QueryServiceEndpointsResponse{
			Msg: "query service endpoints failed: " + err.Error(),
		}
	}
	return &response.QueryServiceEndpointsResponse{
		Results: endpoints,
	}
}

func (s *service) getServiceEndpointsByApo(ctx core.Context, req *request.QueryServiceEndpointsRequest) *response.QueryServiceEndpointsResponse {
	// Get the list of service Endpoint
	endpoints, err := s.promRepo.GetServiceEndPointList(ctx, req.StartTime, req.EndTime, req.ServiceName)
	if err != nil {
		return &response.QueryServiceEndpointsResponse{
			Msg: "query service endpoints failed: " + err.Error(),
		}
	}

	return &response.QueryServiceEndpointsResponse {
		Results: endpoints,
	}
}
