// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type PodMapRequest struct {
	StartTime int64  `form:"startTime" json:"startTime" binding:"required,min=0"`         // query start time
	EndTime   int64  `form:"endTime" json:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Namespace string `form:"namespace"`
	Workload  string `form:"workload"`
	Protocol  string `form:"protocol"`
}

type SpanSegmentMetricsRequest struct {
	TraceId string `form:"traceId" binding:"required"`
	SpanId  string `form:"spanId"`
}
