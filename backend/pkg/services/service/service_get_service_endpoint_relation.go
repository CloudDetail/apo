package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
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
		Parents: parents,
		Current: clickhouse.TopologyNode{
			Service:  req.Service,
			Endpoint: req.Endpoint,
			IsTraced: true,
		},
		ChildRelation: relations,
	}
	return res, nil
}
