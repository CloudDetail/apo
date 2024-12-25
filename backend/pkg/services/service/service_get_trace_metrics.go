// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetTraceMetrics(req *request.GetTraceMetricsRequest) ([]*response.GetTraceMetricsResponse, error) {
	// 获取Trace相关指标
	serviceInstances, err := s.promRepo.GetInstanceList(req.StartTime, req.EndTime, req.Service, req.Endpoint)
	if err != nil {
		return nil, err
	}

	result := make([]*response.GetTraceMetricsResponse, 0)
	for _, instance := range serviceInstances.GetInstances() {
		// 日志告警基于Instance分组查询
		logs, err := s.promRepo.QueryLogCountByInstanceId(instance, req.StartTime, req.EndTime, req.Step)
		if err != nil {
			return nil, err
		}
		// P90延时 基于实例查询
		p90, err := s.promRepo.QueryInstanceP90(req.StartTime, req.EndTime, req.Step, req.Endpoint, instance)
		if err != nil {
			return nil, err
		}
		// 错误率 基于实例查询
		errorRate, err := s.promRepo.QueryInstanceErrorRate(req.StartTime, req.EndTime, req.Step, req.Endpoint, instance)
		if err != nil {
			return nil, err
		}
		// 只显示有数据的实例列表
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
