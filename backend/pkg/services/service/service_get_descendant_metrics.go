// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetDescendantMetrics(ctx core.Context, req *request.GetDescendantMetricsRequest) ([]response.GetDescendantMetricsResponse, error) {
	// Query all descendant nodes
	nodes, err := s.chRepo.ListDescendantNodes(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(nodes.Nodes) == 0 {
		return make([]response.GetDescendantMetricsResponse, 0), nil
	}

	// In addition to descendant nodes, the current node needs to be supplemented
	nodes.AddServerNode(fmt.Sprintf("%s.%s", req.Service, req.Endpoint), req.Service, req.Endpoint, true)

	serverResult, err := s.promRepo.QueryRangePercentile(ctx, req.StartTime, req.EndTime, req.Step, nodes)
	if err != nil {
		return nil, err
	}
	dbResult, err := s.promRepo.QueryDbRangePercentile(ctx, req.StartTime, req.EndTime, req.Step, nodes)
	if err != nil {
		return nil, err
	}
	externalResult, err := s.promRepo.QueryExternalRangePercentile(ctx, req.StartTime, req.EndTime, req.Step, nodes)
	if err != nil {
		return nil, err
	}
	mqResult, err := s.promRepo.QueryMqRangePercentile(ctx, req.StartTime, req.EndTime, req.Step, nodes)
	if err != nil {
		return nil, err
	}
	result := make([]response.GetDescendantMetricsResponse, 0)
	result = append(result, serverResult...)
	result = append(result, dbResult...)
	result = append(result, externalResult...)
	result = append(result, mqResult...)
	return result, nil
}
