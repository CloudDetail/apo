// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetFaultLogPageListRequest struct {
	StartTime   int64    `json:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64    `json:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Service     []string `json:"service"`                                      // query service name
	Namespaces  []string `json:"namespaces"`
	Instance    string   `json:"instance"`    // instance name
	NodeName    string   `json:"nodeName"`    // hostname
	ContainerId string   `json:"containerId"` // container name
	Pid         uint32   `json:"pid"`         // process number
	TraceId     string   `json:"traceId"`     // TraceId
	PageNum     int      `json:"pageNum"`     // page
	PageSize    int      `json:"pageSize"`    // display number per page
}

type GetFaultLogContentRequest struct {
	ServiceName string `json:"serviceName"`
	InstanceId  string `json:"instanceId"`
	TraceId     string `json:"traceId"`
	StartTime   uint64 `json:"startTime"`
	EndTime     uint64 `json:"endTime"`
	EndPoint    string `json:"endpoint"`
	PodName     string `json:"podName"`
	ContainerId string `json:"containerId"`
	NodeName    string `json:"nodeName"`
	Pid         uint32 `json:"pid"`
	SourceFrom  string `json:"sourceFrom"` // log data source
}
