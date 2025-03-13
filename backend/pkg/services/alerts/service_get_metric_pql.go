// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetMetricPQL(ctx core.Context) (*response.GetMetricPQLResponse, error) {
	alertMetrics, err := s.dbRepo.ListQuickAlertRuleMetric(ctx.LANG())
	if err != nil {
		return nil, err
	}
	return &response.GetMetricPQLResponse{
		AlertMetricsData: alertMetrics,
	}, nil
}
