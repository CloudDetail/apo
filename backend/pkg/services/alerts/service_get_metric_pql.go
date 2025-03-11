// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetMetricPQL(req *request.GetMetricPQLRequest) (*response.GetMetricPQLResponse, error) {
	alertMetrics, err := s.dbRepo.ListQuickAlertRuleMetric(req.Language)
	if err != nil {
		return nil, err
	}
	return &response.GetMetricPQLResponse{
		AlertMetricsData: alertMetrics,
	}, nil
}
