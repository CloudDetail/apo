// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

func (s *service) GetServiceName(ctx core.Context, req *request.QueryServiceNameRequest) *response.QueryServiceNameResponse {
	filter := prometheus.NewFilter()
	filter.EqualIfNotEmpty("cluster_id", req.Cluster)
	filter.EqualIfNotEmpty("node_name", req.Tags.NodeName)
	filter.EqualIfNotEmpty("pod_name", req.Tags.PodName)
	filter.EqualIfNotEmpty("container_id", req.Tags.ContainerId)
	services, err := s.promRepo.GetDataplaneServiceList(ctx, req.StartTime, req.EndTime, filter.String())
	if err != nil {
		return &response.QueryServiceNameResponse{
			Msg: "query service name failed: " + err.Error(),
		}
	}
	if len(services) == 0 {
		return &response.QueryServiceNameResponse{
			Msg: "Data Not Found",
		}
	}
	// if len(services) > 1 {
	// 	return &response.QueryServiceNameResponse{
	// 		Msg: fmt.Sprintf("more than one service name[%v] is found by instance", services),
	// 	}
	// }
	return &response.QueryServiceNameResponse{
		Result: services[0].Name,
	}
}

func (s *service) getServiceNameByApo(ctx core.Context, req *request.QueryServiceNameRequest) *response.QueryServiceNameResponse {
	filterKVs := make([]string, 0)
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
	if len(services) > 1 {
		return &response.QueryServiceNameResponse{
			Msg: fmt.Sprintf("more than one service name[%v] is found by instance", services),
		}
	}
	return &response.QueryServiceNameResponse{
		Result: services[0],
	}
}
