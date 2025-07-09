// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetTracePageListRequest struct {
	GroupID     int64    `json:"groupId,omitempty"`                            // Data group id
	StartTime   int64    `json:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64    `json:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Service     []string `json:"service"`                                      // query service name
	Namespace   []string `json:"namespace"`
	EndPoint    string   `json:"endpoint"`    // query Endpoint
	Instance    string   `json:"instance"`    // instance name
	NodeName    string   `json:"nodeName"`    // hostname
	ContainerId string   `json:"containerId"` // container name
	Pid         uint32   `json:"pid"`         // process number
	TraceId     string   `json:"traceId"`     // TraceId
	PageNum     int      `json:"pageNum"`     // page
	PageSize    int      `json:"pageSize"`
	ClusterIDs  []string `json:"clusterIds" form:"clusterIds"`

	Filters []*ComplexSpanTraceFilter `json:"filters"` // filter
}

type GetOnOffCPURequest struct {
	PID       uint32 `form:"pid" binding:"required"`
	NodeName  string `form:"nodeName" binding:"required"`
	StartTime int64  `form:"startTime" binding:"required"`
	EndTime   int64  `form:"endTime" binding:"required"`
}

type GetSingleTraceInfoRequest struct {
	TraceID string `form:"traceId" binding:"required"`
}

type GetFlameDataRequest struct {
	SampleType string `json:"sampleType" form:"sampleType" binding:"required"`
	PID        int64  `json:"pid" form:"pid" binding:"required"`
	TID        int64  `json:"tid" form:"tid" binding:"required"`
	NodeName   string `json:"nodeName" form:"nodeName"`
	SpanID     string `json:"spanId" form:"spanId" binding:"required"`
	TraceID    string `json:"traceId" form:"traceId" binding:"required"`
	StartTime  int64  `json:"startTime" form:"startTime" binding:"required"`
	EndTime    int64  `json:"endTime" form:"endTime" binding:"required,gtfield=StartTime"`
}

type GetProcessFlameGraphRequest struct {
	// Limit the minimum total to be displayed by the node
	MaxNodes   int64  `json:"maxNodes" form:"maxNodes"`
	StartTime  int64  `json:"startTime" form:"startTime" binding:"required"`
	EndTime    int64  `json:"endTime" form:"endTime" binding:"required,gtfield=StartTime"`
	PID        int64  `json:"pid" form:"pid" binding:"required"`
	NodeName   string `json:"nodeName" form:"nodeName"`
	SampleType string `json:"sampleType" form:"sampleType" binding:"required"`
}
