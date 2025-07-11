// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServices(ctx core.Context, req *request.QueryServicesRequest) *response.QueryServicesResponse {
	// Get the list of service
	services, err := s.chRepo.QueryServices(ctx, req.StartTime, req.EndTime, req.Cluster)
	if err != nil {
		return &response.QueryServicesResponse{
			Msg: "query service failed: " + err.Error(),
		}
	}
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
