// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceEndpointTopology(req *request.GetServiceEndpointTopologyRequest) (*response.GetServiceEndpointTopologyResponse, error) {
	// Query all upstream nodes
	parents, err := s.chRepo.ListParentNodes(req)
	if err != nil {
		return nil, err
	}

	// Query all downstream nodes
	children, err := s.chRepo.ListChildNodes(req)
	if err != nil {
		return nil, err
	}

	res := &response.GetServiceEndpointTopologyResponse{
		Parents:  parents.GetNodes(),
		Current:  model.NewServerNode(req.Service, req.Endpoint, true),
		Children: children.GetNodes(),
	}
	return res, nil
}
