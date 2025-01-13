// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type GetSQLMetricsResponse struct {
	Pagination          model.Pagination     `json:"pagination"`
	SQLOperationDetails []SQLOperationDetail `json:"sqlOperationDetails"`
}

type SQLOperationDetail struct {
	prometheus.SQLKey

	Latency   TempChartObject `json:"latency"`
	ErrorRate TempChartObject `json:"errorRate"`
	// FIXME Tps name is tps, actual requests per minute
	Tps TempChartObject `json:"tps"`
}
