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
	SummaryOnly            bool   `form:"summaryOnly"`
}

type ListSystemHealthSystemsRequest struct {
	RelativeTime string `form:"relativeTime"`
	StartTime    int64  `form:"startTime"`
	EndTime      int64  `form:"endTime"`
	Page         int    `form:"page"`
	PageSize     int    `form:"pageSize"`
	Keyword      string `form:"keyword"`
	Status       string `form:"status"`
	Sort         string `form:"sort"`
}

type ListSystemHealthSystemServicesRequest struct {
	GroupID int64 `uri:"groupId" binding:"required"`

	RelativeTime string `form:"relativeTime"`
	StartTime    int64  `form:"startTime"`
	EndTime      int64  `form:"endTime"`
	Page         int    `form:"page"`
	PageSize     int    `form:"pageSize"`
	Keyword      string `form:"keyword"`
	Status       string `form:"status"`
	Sort         string `form:"sort"`
}

type ListSystemHealthPendingIssuesRequest struct {
	RelativeTime string `form:"relativeTime"`
	StartTime    int64  `form:"startTime"`
	EndTime      int64  `form:"endTime"`
	Page         int    `form:"page"`
	PageSize     int    `form:"pageSize"`
	Keyword      string `form:"keyword"`
	Status       string `form:"status"`
	GroupID      int64  `form:"groupId"`
}

type GetSystemHealthNewAbnormalMetricsRequest struct {
	RelativeTime string `form:"relativeTime"`
	StartTime    int64  `form:"startTime"`
	EndTime      int64  `form:"endTime"`
}

type GetSystemHealthTrendRequest struct {
	Range     string `form:"range"`
	GroupID   int64  `form:"groupId"`
	Judgement string `form:"judgement"`
	View      string `form:"view"`
}

type GetSystemHealthSystemsTrendRequest struct {
	Range     string  `json:"range"`
	Judgement string  `json:"judgement"`
	GroupIDs  []int64 `json:"groupIds"`
}

type UpdateSystemHealthConfigRequest struct {
	AbnormalScoreThreshold        int `json:"abnormalScoreThreshold"`
	SevereAbnormalScoreThreshold  int `json:"severeAbnormalScoreThreshold" binding:"required,min=1,max=99"`
	GeneralAbnormalScoreThreshold int `json:"generalAbnormalScoreThreshold" binding:"required,min=2,max=100"`
}

type GetSystemHealthSystemDetailRequest struct {
	GroupID int64 `uri:"groupId" binding:"required"`

	RelativeTime    string `form:"relativeTime"`
	StartTime       int64  `form:"startTime"`
	EndTime         int64  `form:"endTime"`
	IncludeServices *bool  `form:"includeServices"`
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
