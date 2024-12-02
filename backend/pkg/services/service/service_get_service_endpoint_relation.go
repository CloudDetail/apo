package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetServiceEndpointRelation(req *request.GetServiceEndpointRelationRequest) (*response.GetServiceEndpointRelationResponse, error) {
	// 查询所有上游节点
	parents, err := s.chRepo.ListParentNodes(&req.GetServiceEndpointTopologyRequest)
	if err != nil {
		return nil, err
	}

	var relations []*model.ToplogyRelation

	// TODO 下次更新移除,兼容旧前端
	if req.RemoveClientCall {
		relations, err = s.chRepo.ListDescendantRelationsWithoutEdge(&req.GetServiceEndpointTopologyRequest)
	} else {
		// 查询所有下游节点的调用关系列表
		relations, err = s.chRepo.ListDescendantRelations(&req.GetServiceEndpointTopologyRequest)
	}

	if err != nil {
		return nil, err
	}

	res := &response.GetServiceEndpointRelationResponse{
		Parents:       parents.GetNodes(),
		Current:       model.NewServerNode(req.Service, req.Endpoint, true),
		ChildRelation: relations,
	}

	// TODO 下次更新移除,兼容旧前端
	if req.WithTopologyLevel {
		if req.EntryService == "" || req.EntryEndpoint == "" {
			req.EntryService = req.Service
			req.EntryEndpoint = req.Endpoint
		}

		// 构造邻接矩阵
		endpointMatrix := newEndpointDepthMatrix(req.EntryService, req.EntryEndpoint, relations)
		if req.WithTopologyLevel {
			// 下游接口评估非循环下的最大深度
			res.TopologyLevels = endpointMatrix.EndpointsLevel()
		}
	}
	return res, nil
}
