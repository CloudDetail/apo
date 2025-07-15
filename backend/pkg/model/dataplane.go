// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"slices"
	"strings"
	"time"
)

type Service struct {
	ClusterId string `json:"clusterId"`
	Id        string `json:"id"`
	Name      string `json:"name"`
	Source    string `json:"source"`
}

type RedCharts struct {
	Count      map[int64]int64 `json:"count"`
	ErrorCount map[int64]int64 `json:"errorCount"`
	Duration   map[int64]int64 `json:"duration"`
}

type RedMetricValue struct {
	Count      int64 `json:"count"`
	ErrorCount int64 `json:"errorCount"`
	Duration   int64 `json:"duration"`
}

type ServiceToplogy struct {
	ParentService string `ch:"parent_service" json:"parentService"`
	ParentType    string `ch:"parent_type" json:"parentType"`
	ChildService  string `ch:"child_service" json:"childService"`
	ChildType     string `ch:"child_type" json:"childType"`
}

type ApmServiceInstance struct {
	ServiceName string   `json:"serviceName"`
	HostName    string   `json:"hostName"`
	ProcessId   string   `json:"processId"`
	ContainerId string   `json:"containerId"`
	Ips         []string `json:"ips"`
	NodeIp      string   `json:"nodeIp"`
	NodeName    string   `json:"nodeName"`
}

type AppInfo struct {
	Timestamp       time.Time         `ch:"timestamp" json:"timestamp"`
	StartTime       uint64            `ch:"start_time" json:"startTime"`
	AgentInstanceId string            `ch:"agent_instance_id" json:"agentInstanceId"`
	HostPid         uint32            `ch:"host_pid" json:"hostPid"`
	ContainerPid    uint32            `ch:"container_pid" json:"containerPid"`
	ContainerId     string            `ch:"container_id" json:"containerId"`
	Labels          map[string]string `ch:"labels" json:"labels"`
}

func (app *AppInfo) CheckIps(ips []string) bool {
	expectIps := strings.Split(app.Labels["ip"], ",")
	for _, ip := range ips {
		if !slices.Contains(expectIps, ip) {
			return false
		}
	}
	return true
}
