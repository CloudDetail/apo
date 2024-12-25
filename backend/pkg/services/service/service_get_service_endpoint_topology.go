// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceEndpointTopology(req *request.GetServiceEndpointTopologyRequest) (*response.GetServiceEndpointTopologyResponse, error) {
	// 查询所有上游节点
	parents, err := s.chRepo.ListParentNodes(req)
	if err != nil {
		return nil, err
	}

	// 查询所有下游节点
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
