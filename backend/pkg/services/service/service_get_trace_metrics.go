// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetTraceMetrics(ctx_core core.Context, req *request.GetTraceMetricsRequest) ([]*response.GetTraceMetricsResponse, error) {
	// Get Trace related metrics
	serviceInstances, err := s.promRepo.GetInstanceList(req.StartTime, req.EndTime, req.Service, req.Endpoint)
	if err != nil {
		return nil, err
	}

	result := make([]*response.GetTraceMetricsResponse, 0)
	for _, instance := range serviceInstances.GetInstances() {
		// Log alarm query based on Instance grouping
		logs, err := s.promRepo.QueryLogCountByInstanceId(instance, req.StartTime, req.EndTime, req.Step)
		if err != nil {
			return nil, err
		}
		// P90 delayed instance-based query
		p90, err := s.promRepo.QueryInstanceP90(req.StartTime, req.EndTime, req.Step, req.Endpoint, instance)
		if err != nil {
			return nil, err
		}
		// Error rate for instance-based queries
		errorRate, err := s.promRepo.QueryInstanceErrorRate(req.StartTime, req.EndTime, req.Step, req.Endpoint, instance)
		if err != nil {
			return nil, err
		}
		// Display only the list of instances with data
		if exist_metrics(logs) || exist_metrics(p90) || exist_metrics(errorRate) {
			metricResponse := &response.GetTraceMetricsResponse{
				Name:        instance.GetInstanceId(),
				ContainerId: instance.ContainerId,
				NodeName:    instance.NodeName,
				Pid:         instance.Pid,
				Logs:        logs,
				Latency:     p90,
				ErrorRate:   errorRate,
			}
			result = append(result, metricResponse)
		}
	}
	return result, nil
}
