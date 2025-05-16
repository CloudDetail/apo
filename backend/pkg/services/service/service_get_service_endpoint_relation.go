// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceEndpointRelation(ctx core.Context, req *request.GetServiceEndpointRelationRequest) (*response.GetServiceEndpointRelationResponse, error) {
	// Query all upstream nodes
	parents, err := s.chRepo.ListParentNodes(req)
	if err != nil {
		return nil, err
	}

	// Query the calling relationship list of all downstream nodes
	relations, err := s.chRepo.ListDescendantRelations(req)
	if err != nil {
		return nil, err
	}

	res := &response.GetServiceEndpointRelationResponse{
		Parents:       parents.GetNodes(),
		Current:       model.NewServerNode(req.Service, req.Endpoint, true),
		ChildRelation: relations,
	}
	return res, nil
}
