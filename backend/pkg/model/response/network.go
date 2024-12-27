// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

type PodMapResponse struct {
	Columns []string `json:"columns"`
	Schemas []struct {
		LabelType string `json:"label_type"`
		PreAs     string `json:"pre_as"`
		Type      int    `json:"type"`
		Unit      string `json:"unit"`
		ValueType string `json:"value_type"`
	} `json:"schemas"`
	Values [][]interface{} `json:"values"`
}

type SegmentLatency struct {
	// 在客户端进程处采集到的请求和响应时间
	ClientProcess Duration `json:"clientProcess"`
	// 在客户端网卡处采集到的请求和响应时间
	ClientNic Duration `json:"clientNic"`
	// 在客户端主机网卡处采集到的请求和响应时间
	ClientK8sNodeNic Duration `json:"clientK8SNodeNic"`
	// 在服务端主机网卡处采集到的请求和响应时间
	ServerK8sNodeNic Duration `json:"serverK8SNodeNic"`
	// 在服务端网卡处采集到的请求和响应时间
	ServerNic Duration `json:"serverNic"`
	// 在服务端进程处采集到的请求和响应时间
	ServerProcess Duration `json:"serverProcess"`
}

type Duration struct {
	// 请求网络包时间戳，单位微秒
	StartTime int64 `json:"startTime"`
	// 响应网络包时间戳，单位微秒
	EndTime int64 `json:"endTime"`
	// 响应延时
	ResponseDuration uint64 `json:"responseDuration"`
}

type SpanSegmentMetricsResponse map[string]*SegmentLatency
