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
	// Request and response times collected at the client process
	ClientProcess Duration `json:"clientProcess"`
	// Request and response time collected at the client NIC
	ClientNic Duration `json:"clientNic"`
	// Request and response time collected at the client host NIC
	ClientK8sNodeNic Duration `json:"clientK8SNodeNic"`
	// Request and response time collected at the server host NIC
	ServerK8sNodeNic Duration `json:"serverK8SNodeNic"`
	// Request and response time collected at the server NIC
	ServerNic Duration `json:"serverNic"`
	// Request and response time collected at the server process
	ServerProcess Duration `json:"serverProcess"`
}

type Duration struct {
	// The timestamp of the request network packet, in microseconds.
	StartTime int64 `json:"startTime"`
	// Timestamp of the response network packet, in microseconds
	EndTime int64 `json:"endTime"`
	// Response delay
	ResponseDuration uint64 `json:"responseDuration"`
}

type SpanSegmentMetricsResponse map[string]*SegmentLatency
