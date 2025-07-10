// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) GetServiceEndpointTopology(ctx core.Context, req *request.GetServiceEndpointTopologyRequest) (*response.GetServiceEndpointTopologyResponse, error) {
	// Query all upstream nodes
	parents, err := s.chRepo.ListParentNodes(ctx, req)
	if err != nil {
		return nil, err
	}

	parents, err = common.MarkTopologyNodeInGroup(ctx, s.dbRepo, req.GroupID, parents)
	if err != nil {
		return nil, err
	}

	// Query all downstream nodes
	children, err := s.chRepo.ListChildNodes(ctx, req)
	if err != nil {
		return nil, err
	}

	children, err = common.MarkTopologyNodeInGroup(ctx, s.dbRepo, req.GroupID, children)
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
