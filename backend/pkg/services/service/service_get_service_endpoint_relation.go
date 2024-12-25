// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceEndpointRelation(req *request.GetServiceEndpointRelationRequest) (*response.GetServiceEndpointRelationResponse, error) {
	// 查询所有上游节点
	parents, err := s.chRepo.ListParentNodes(req)
	if err != nil {
		return nil, err
	}

	// 查询所有下游节点的调用关系列表
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
