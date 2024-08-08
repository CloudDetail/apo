package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetDescendantMetrics(req *request.GetDescendantMetricsRequest) ([]response.GetDescendantMetricsResponse, error) {
	// 查询所有子孙节点
	nodes, err := s.chRepo.ListDescendantNodes(req)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return make([]response.GetDescendantMetricsResponse, 0), nil
	}

	svcs := make([]string, 0)
	endpoints := make([]string, 0)
	for _, node := range nodes {
		if node.IsTraced {
			svcs = append(svcs, node.Service)
			endpoints = append(endpoints, node.Endpoint)
		}
	}
	// 除了子孙节点，还需补充当前节点
	svcs = append(svcs, req.Service)
	endpoints = append(endpoints, req.Endpoint)

	return s.promRepo.QueryRangePercentile(req.StartTime, req.EndTime, req.Step, svcs, endpoints)
}
