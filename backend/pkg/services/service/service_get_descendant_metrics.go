package service

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetDescendantMetrics(req *request.GetDescendantMetricsRequest) ([]response.GetDescendantMetricsResponse, error) {
	// 查询所有子孙节点
	nodes, err := s.chRepo.ListDescendantNodes(req)
	if err != nil {
		return nil, err
	}
	if len(nodes.Nodes) == 0 {
		return make([]response.GetDescendantMetricsResponse, 0), nil
	}

	// 除了子孙节点，还需补充当前节点
	nodes.AddServerNode(fmt.Sprintf("%s.%s", req.Service, req.Endpoint), req.Service, req.Endpoint, true)

	serverResult, err := s.promRepo.QueryRangePercentile(req.StartTime, req.EndTime, req.Step, nodes)
	if err != nil {
		return nil, err
	}
	dbResult, err := s.promRepo.QueryDbRangePercentile(req.StartTime, req.EndTime, req.Step, nodes)
	if err != nil {
		return nil, err
	}
	externalResult, err := s.promRepo.QueryExternalRangePercentile(req.StartTime, req.EndTime, req.Step, nodes)
	if err != nil {
		return nil, err
	}
	mqResult, err := s.promRepo.QueryMqRangePercentile(req.StartTime, req.EndTime, req.Step, nodes)
	if err != nil {
		return nil, err
	}
	result := make([]response.GetDescendantMetricsResponse, 0)
	result = append(result, serverResult...)
	result = append(result, dbResult...)
	result = append(result, externalResult...)
	result = append(result, mqResult...)
	return result, nil
}
