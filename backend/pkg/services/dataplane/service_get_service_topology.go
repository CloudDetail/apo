// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceTopology(ctx core.Context, req *request.QueryTopologyRequest) *response.QueryTopologyResponse {
	serviceTopologys, err := s.chRepo.ListServiceTopologys(ctx, req)
	if err != nil {
		return &response.QueryTopologyResponse{
			Msg: "query service topology failed: " + err.Error(),
		}
	}
	if len(serviceTopologys.Nodes) == 0 {
		return s.queryServiceTopologyByApi(ctx, req)
	}

	results := make([]*model.ServiceToplogyNode, 0)
	for _, node := range serviceTopologys.Nodes {
		results = append(results, node)
	}
	return &response.QueryTopologyResponse{
		Results: results,
	}
}

func (s *service) queryServiceTopologyByApi(ctx core.Context, req *request.QueryTopologyRequest) *response.QueryTopologyResponse {
	return &response.QueryTopologyResponse{
		Msg: "Data not found",
	}
}
