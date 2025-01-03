// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetMetricPQL() (*response.GetMetricPQLResponse, error) {
	alertMetrics, err := s.dbRepo.ListQuickAlertRuleMetric()
	if err != nil {
		return nil, err
	}
	return &response.GetMetricPQLResponse{
		AlertMetricsData: alertMetrics,
	}, nil
}
