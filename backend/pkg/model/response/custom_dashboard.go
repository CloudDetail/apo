// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model/customdashboard"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

type DashboardListResponse struct {
	Dashboards []customdashboard.Dashboard `json:"dashboards"`
	CanCreate  bool                        `json:"canCreate"`
}

type DashboardValidateQueryResponse struct {
	Valid    bool   `json:"valid"`
	Message  string `json:"message,omitempty"`
	Position int    `json:"position"`
}

type DashboardQueryResponse struct {
	Status      string                    `json:"status"`
	Message     string                    `json:"message,omitempty"`
	Position    int                       `json:"position"`
	Query       string                    `json:"query"`
	Series      []prometheus.MetricResult `json:"series"`
	AuditID     uint                      `json:"auditId"`
	Truncated   bool                      `json:"truncated"`
	SeriesLimit *DashboardSeriesLimitInfo `json:"seriesLimit,omitempty"`
	DataScope   string                    `json:"dataScope"`
}

type DashboardSeriesLimitInfo struct {
	Limited bool `json:"limited"`
	Total   int  `json:"total"`
	Shown   int  `json:"shown"`
	Limit   int  `json:"limit"`
}
