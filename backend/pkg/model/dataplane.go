// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"slices"
	"strconv"
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

func (app *AppInfo) GetInstanceName(serviceName string) string {
	if podName, ok := app.Labels["pod_name"]; ok && podName != "" {
		return podName
	}
	svcName := app.Labels["service_name"]
	if svcName == "" {
		svcName = serviceName
	}
	if app.ContainerId != "" {
		return svcName + "@" + app.Labels["node_name"] + "@" + app.ContainerId
	}
	return svcName + "@" + app.Labels["node_name"] + "@" + strconv.FormatInt(int64(app.HostPid), 10)
}

func (app *AppInfo) GetSource() string {
	if ruleId, ok := app.Labels["rule_id"]; ok && ruleId != "" {
		return "Rule"
	}
	if source, ok := app.Labels["source"]; ok && source != "" {
		return "Source-" + app.Labels["source"]
	}
	return ""
}

func (app *AppInfo) GetService() string {
	serviceName := app.Labels["service_name"]
	return serviceName
}

type MatchServiceInstance struct {
	ServiceName  string `json:"serviceName"`
	InstanceName string `json:"instanceName"`
	Source       string `json:"source"`
}

func NewMatchServiceInstance(serviceName string, app *AppInfo) *MatchServiceInstance {
	return &MatchServiceInstance{
		ServiceName:  app.GetService(),
		InstanceName: app.GetInstanceName(serviceName),
		Source:       app.GetSource(),
	}
}

type CheckDataSourceRequest struct {
	Datasource   string   `json:"dataSource" binding:"required"` // query Datasource
	Attributes   string   `json:"attributes" binding:"required"` // query attributes
	Capabilities []string `json:"enabledCapabilities"`           // query capabilities
}

type CheckDataSourceResponse struct {
	Message string `json:"message,omitempty"`
	Success bool   `json:"success"`
}
