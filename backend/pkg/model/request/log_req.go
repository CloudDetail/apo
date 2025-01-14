// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetFaultLogPageListRequest struct {
	StartTime   int64    `json:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime     int64    `json:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Service     []string `json:"service"`                                      // 查询服务名
	Namespaces  []string `json:"namespaces"`
	Instance    string   `json:"instance"`    // 实例名
	NodeName    string   `json:"nodeName"`    // 主机名
	ContainerId string   `json:"containerId"` // 容器名
	Pid         uint32   `json:"pid"`         // 进程号
	TraceId     string   `json:"traceId"`     // TraceId
	PageNum     int      `json:"pageNum"`     // 第几页
	PageSize    int      `json:"pageSize"`    // 每页显示条数
}

type GetFaultLogContentRequest struct {
	ServiceName string `json:"serviceName"` // unused
	InstanceId  string `json:"instanceId"`
	TraceId     string `json:"traceId"`
	StartTime   uint64 `json:"startTime"`
	EndTime     uint64 `json:"endTime"`
	EndPoint    string `json:"endpoint"`
	PodName     string `json:"podName"`
	ContainerId string `json:"containerId"`
	NodeName    string `json:"nodeName"`
	Pid         uint32 `json:"pid"`
	SourceFrom  string `json:"sourceFrom"` // 日志数据源
}
