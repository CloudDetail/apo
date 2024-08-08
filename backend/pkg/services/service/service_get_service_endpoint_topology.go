package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
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
		Parents: parents,
		Current: clickhouse.TopologyNode{
			Service:  req.Service,
			Endpoint: req.Endpoint,
			IsTraced: true,
		},
		Children: children,
	}
	return res, nil
}
