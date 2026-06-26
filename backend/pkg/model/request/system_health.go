// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetSystemHealthReportRequest struct {
	GroupID      int64  `form:"groupId"`
	RelativeTime string `form:"relativeTime"`
	StartTime    int64  `form:"startTime"`
	EndTime      int64  `form:"endTime"`
}

type GetSystemHealthOverviewRequest struct {
	RelativeTime           string `form:"relativeTime"`
	StartTime              int64  `form:"startTime"`
	EndTime                int64  `form:"endTime"`
	SkipNewAbnormalMetrics bool   `form:"skipNewAbnormalMetrics"`
}

type GetSystemHealthNewAbnormalMetricsRequest struct {
	RelativeTime string `form:"relativeTime"`
	StartTime    int64  `form:"startTime"`
	EndTime      int64  `form:"endTime"`
}

type UpdateSystemHealthConfigRequest struct {
	AbnormalScoreThreshold int `json:"abnormalScoreThreshold" binding:"required,min=1,max=100"`
}

type GetSystemHealthSystemDetailRequest struct {
	GroupID int64 `uri:"groupId" binding:"required"`

	RelativeTime string `form:"relativeTime"`
	StartTime    int64  `form:"startTime"`
	EndTime      int64  `form:"endTime"`
}

type AnalyzeSystemHealthPendingIssueRequest struct {
	GroupID     int64  `json:"groupId" binding:"required"`
	ServiceName string `json:"serviceName" binding:"required"`
	MetricKey   string `json:"metricKey" binding:"required"`

	RelativeTime string `json:"relativeTime"`
	StartTime    int64  `json:"startTime"`
	EndTime      int64  `json:"endTime"`
}

type ListSystemHealthReportRecordsRequest struct {
	GroupID  int64 `form:"groupId"`
	Page     int   `form:"page"`
	PageSize int   `form:"pageSize"`
}

type GetSystemHealthReportRecordRequest struct {
	RecordID string `uri:"recordId"`
}
