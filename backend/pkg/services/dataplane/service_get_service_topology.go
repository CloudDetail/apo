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
	realTopologies, err := s.chRepo.QueryRealtimeServiceTopology(ctx, req.StartTime, req.EndTime, req.Cluster)
	if err != nil {
		return &response.QueryTopologyResponse{
			Msg: "query realtime topology failed: " + err.Error(),
		}
	}
	staticTopologies, err := s.dbRepo.ListCustomServiceTopology(ctx)
	if err != nil {
		return &response.QueryTopologyResponse{
			Msg: "query custom topology failed: " + err.Error(),
		}
	}
	nodes := make(map[string]*model.ServiceToplogyNode)
	for _, realTopology := range realTopologies {
		var (
			parentNode *model.ServiceToplogyNode
			childNode *model.ServiceToplogyNode
			ok bool
		)
		if realTopology.ParentService != "" {
			parentNode, ok = nodes[realTopology.ParentService]
			if !ok {
				parentNode = model.NewServiceToplogyNode(realTopology.ParentService, realTopology.ParentType, false)
				nodes[realTopology.ParentService] = parentNode 
			}
		}
		if realTopology.ChildService != "" {
			childNode, ok = nodes[realTopology.ChildService]
			if !ok {
				childNode = model.NewServiceToplogyNode(realTopology.ChildService, realTopology.ChildType, false)
				nodes[realTopology.ChildService] = childNode 
			}
		}
		if parentNode != nil && childNode != nil {
			parentNode.AddChild(childNode)
		}
	}

	for _, staticTopology := range staticTopologies {
		if staticTopology.ExpireTime == 0 || staticTopology.ExpireTime >= req.EndTime {
			parentNode, ok := nodes[staticTopology.LeftNode]
			if !ok {
				parentNode = model.NewServiceToplogyNode(staticTopology.LeftNode, staticTopology.LeftType, true)
				nodes[staticTopology.LeftNode] = parentNode 
			}
			childNode, ok := nodes[staticTopology.RightNode]
			if !ok {
				childNode = model.NewServiceToplogyNode(staticTopology.RightNode, staticTopology.RightType, true)
				nodes[staticTopology.RightNode] = childNode 
			}
			parentNode.AddChild(childNode)
		}
	}

	results := make([]*model.ServiceToplogyNode, 0)
	for _, node := range nodes {
		results = append(results, node)
	}
	return &response.QueryTopologyResponse{
		Results: results,
	}
}

func (s *service) getServiceTopologyByApo(ctx core.Context, req *request.QueryTopologyRequest) *response.QueryTopologyResponse {
	serviceTopologys, err := s.chRepo.ListServiceTopologys(ctx, req)
	if err != nil {
		return &response.QueryTopologyResponse{
			Msg: "query service topology failed: " + err.Error(),
		}
	}
	results := make([]*model.ServiceToplogyNode, 0)
	for _, node := range serviceTopologys.Nodes {
		results = append(results, node)
	}
	return &response.QueryTopologyResponse{
		Results: results,
	}
}