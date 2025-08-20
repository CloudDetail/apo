// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServiceName(ctx core.Context, req *request.QueryServiceNameRequest) *response.QueryServiceNameResponse {
	services, err := s.chRepo.GetDataplaneServiceList(ctx, req)
	if err != nil {
		return &response.QueryServiceNameResponse{
			Msg: "query service name failed: " + err.Error(),
		}
	}
	if len(services) > 0 {
		return &response.QueryServiceNameResponse{
			Result: services[0].Name,
		}
	}
	return s.getServiceNameByApo(ctx, req)
}

func (s *service) getServiceNameByApo(ctx core.Context, req *request.QueryServiceNameRequest) *response.QueryServiceNameResponse {
	filterKVs := make([]string, 0)
	if req.Cluster != "" {
		filterKVs = append(filterKVs, prometheus.ClusterIDPQLFilter, req.Cluster)
	}
	if req.Tags.PodName != "" {
		filterKVs = append(filterKVs, prometheus.PodPQLFilter, req.Tags.PodName)
	}
	if req.Tags.NodeName != "" {
		filterKVs = append(filterKVs, prometheus.NodeNamePQLFilter, req.Tags.NodeName)
	}
	if req.Tags.ContainerId != "" {
		filterKVs = append(filterKVs, prometheus.ContainerIdPQLFilter, req.Tags.ContainerId)
	} else if req.Tags.Pid != "" {
		filterKVs = append(filterKVs, prometheus.PidPQLFilter, req.Tags.Pid)
	}
	services, err := s.promRepo.GetServiceListByFilter(
		ctx,
		time.Unix(req.StartTime/1000000, 0), time.Unix(req.EndTime/1000000, 0),
		filterKVs...,
	)
	if err != nil {
		return &response.QueryServiceNameResponse{
			Msg: "query service name by instance info failed: " + err.Error(),
		}
	}
	if len(services) == 0 {
		return &response.QueryServiceNameResponse{
			Msg: "Data Not Found",
		}
	}
	return &response.QueryServiceNameResponse{
		Result: services[0],
	}
}
