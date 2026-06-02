// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package customdashboard

import (
	"github.com/CloudDetail/apo/backend/pkg/model/customdashboard"
)

func defaultVariables() []customdashboard.Variable {
	return []customdashboard.Variable{
		{Name: "cluster", Label: "Cluster", Type: "cluster", Multi: true, IncludeAll: true, AllValue: ".*"},
		{Name: "namespace", Label: "Namespace", Type: "namespace", Multi: true, IncludeAll: true, AllValue: ".*"},
		{Name: "service", Label: "Service", Type: "service", Multi: true, IncludeAll: true, AllValue: ".*"},
		{Name: "endpoint", Label: "Endpoint", Type: "endpoint", Multi: true, IncludeAll: true, AllValue: ".*"},
		{Name: "pod", Label: "Pod", Type: "pod", Multi: true, IncludeAll: true, AllValue: ".*"},
		{Name: "instance", Label: "Instance", Type: "instance", Multi: true, IncludeAll: true, AllValue: ".*"},
	}
}

func blankDashboard() customdashboard.Dashboard {
	return customdashboard.Dashboard{
		Name:       "Untitled Dashboard",
		Visibility: customdashboard.VisibilityEveryone,
		Variables:  defaultVariables(),
		Widgets:    []customdashboard.Widget{},
	}
}

func catalogPresetQueries() []customdashboard.PresetQuery {
	return []customdashboard.PresetQuery{
		{ID: "service-qps", Category: "service", Name: "Request QPS", Description: "Service request throughput", Expression: `sum(rate(span_trace_duration_count{svc_name=~"${service:regex}"}[1m]))`, CompatibleCharts: []string{"line", "area"}, Parameters: serviceParams()},
		{ID: "service-error-rate", Category: "service", Name: "Error Rate", Description: "Service request error ratio", Expression: `sum(rate(span_trace_duration_count{svc_name=~"${service:regex}",is_error="true"}[1m])) / sum(rate(span_trace_duration_count{svc_name=~"${service:regex}"}[1m]))`, CompatibleCharts: []string{"line", "area"}, Parameters: serviceParams()},
		{ID: "service-latency-avg", Category: "service", Name: "Average Latency", Description: "Service average latency", Expression: `sum(rate(span_trace_duration_sum{svc_name=~"${service:regex}"}[5m])) / sum(rate(span_trace_duration_count{svc_name=~"${service:regex}"}[5m]))`, CompatibleCharts: []string{"line", "area"}, Parameters: serviceParams()},
		{ID: "service-latency-p95", Category: "service", Name: "P95 Latency", Description: "Service P95 latency", Expression: `histogram_quantile(0.95, sum(rate(span_trace_duration_bucket{svc_name=~"${service:regex}"}[5m])) by (le))`, CompatibleCharts: []string{"line", "area"}, Parameters: serviceParams()},
		{ID: "service-latency-p99", Category: "service", Name: "P99 Latency", Description: "Service P99 latency", Expression: `histogram_quantile(0.99, sum(rate(span_trace_duration_bucket{svc_name=~"${service:regex}"}[5m])) by (le))`, CompatibleCharts: []string{"line", "area"}, Parameters: serviceParams()},
		{ID: "service-instance-status", Category: "service", Name: "Service Instance Status", Description: "Instance availability for a service", Expression: `sum by (pod) (up{svc_name=~"${service:regex}"})`, CompatibleCharts: []string{"bar", "table"}, Parameters: serviceParams()},
		{ID: "service-slow-endpoints", Category: "service", Name: "Slow Endpoints TopN", Description: "Slowest endpoints by average latency", Expression: `topk(10, sum(rate(span_trace_duration_sum{svc_name=~"${service:regex}"}[5m])) by (content_key) / sum(rate(span_trace_duration_count{svc_name=~"${service:regex}"}[5m])) by (content_key))`, CompatibleCharts: []string{"bar", "table"}, Parameters: serviceParams()},
		{ID: "pod-cpu", Category: "instance", Name: "CPU Usage", Description: "Pod CPU usage", Expression: `sum(rate(container_cpu_usage_seconds_total{pod=~"${pod:regex}"}[5m]))`, CompatibleCharts: []string{"line", "area", "bar"}, Parameters: podParams()},
		{ID: "pod-memory", Category: "instance", Name: "Memory Usage", Description: "Pod memory usage", Expression: `sum(container_memory_working_set_bytes{pod=~"${pod:regex}"})`, CompatibleCharts: []string{"line", "area", "bar"}, Parameters: podParams()},
		{ID: "pod-network-traffic", Category: "instance", Name: "Network Traffic", Description: "Pod receive and transmit traffic", Expression: `sum(rate(container_network_receive_bytes_total{pod=~"${pod:regex}"}[5m])) + sum(rate(container_network_transmit_bytes_total{pod=~"${pod:regex}"}[5m]))`, CompatibleCharts: []string{"line", "area", "bar"}, Parameters: podParams()},
		{ID: "pod-restarts", Category: "instance", Name: "Restart Count", Description: "Pod restart count", Expression: `sum(kube_pod_container_status_restarts_total{pod=~"${pod:regex}"})`, CompatibleCharts: []string{"bar", "table"}, Parameters: podParams()},
		{ID: "node-cpu", Category: "infrastructure", Name: "Node CPU", Description: "Node CPU usage", Expression: `100 - avg(rate(node_cpu_seconds_total{mode="idle"}[5m])) by (instance) * 100`, CompatibleCharts: []string{"line", "area", "table"}, Parameters: nil},
		{ID: "node-memory", Category: "infrastructure", Name: "Node Memory", Description: "Node memory usage", Expression: `(1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100`, CompatibleCharts: []string{"line", "area", "table"}, Parameters: nil},
		{ID: "disk-usage", Category: "infrastructure", Name: "Disk Usage", Description: "Filesystem usage", Expression: `(1 - node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100`, CompatibleCharts: []string{"line", "area", "table"}, Parameters: nil},
		{ID: "node-network-traffic", Category: "infrastructure", Name: "Network Traffic", Description: "Node receive and transmit traffic", Expression: `sum(rate(node_network_receive_bytes_total{device!="lo"}[5m])) by (instance) + sum(rate(node_network_transmit_bytes_total{device!="lo"}[5m])) by (instance)`, CompatibleCharts: []string{"line", "area", "bar", "table"}, Parameters: nil},
		{ID: "pod-status-distribution", Category: "infrastructure", Name: "Pod Status Distribution", Description: "Pod phase distribution", Expression: `sum by (phase) (kube_pod_status_phase{phase=~"Pending|Running|Failed|Succeeded|Unknown"})`, CompatibleCharts: []string{"bar", "table"}, Parameters: nil},
		{ID: "resource-usage-top", Category: "infrastructure", Name: "Resource Usage TopN", Description: "Pods with highest CPU usage", Expression: `topk(10, sum(rate(container_cpu_usage_seconds_total{container!="",pod!=""}[5m])) by (pod))`, CompatibleCharts: []string{"bar", "table"}, Parameters: nil},
	}
}

func serviceParams() []customdashboard.PresetParameter {
	return []customdashboard.PresetParameter{{Name: "service", Label: "Service", Description: "Service variable or concrete service name", Required: true, Default: "$service"}}
}

func podParams() []customdashboard.PresetParameter {
	return []customdashboard.PresetParameter{{Name: "pod", Label: "Instance/Pod", Description: "Pod variable or concrete pod name", Required: true, Default: "$pod"}}
}

func catalogTemplates() []customdashboard.Template {
	blank := blankDashboard()
	service := blankDashboard()
	service.Name = "Service Health"
	service.Description = "Request volume, error rate, latency, instance and resource health."
	service.Widgets = []customdashboard.Widget{
		widget("w-qps", "Request QPS", "line", 0, 0, "service-qps"),
		widget("w-error", "Error Rate", "line", 1, 0, "service-error-rate"),
		widget("w-latency", "P95 Latency", "area", 0, 1, "service-latency-p95"),
		widget("w-latency-p99", "P99 Latency", "area", 1, 1, "service-latency-p99"),
		widget("w-cpu", "CPU Usage", "line", 0, 2, "pod-cpu"),
		widget("w-memory", "Memory Usage", "area", 1, 2, "pod-memory"),
		widget("w-restarts", "Restart Count", "table", 0, 3, "pod-restarts"),
		widget("w-instance-status", "Service Instance Status", "table", 1, 3, "service-instance-status"),
		widget("w-slow-endpoints", "Slow Endpoints TopN", "table", 0, 4, "service-slow-endpoints"),
	}
	infra := blankDashboard()
	infra.Name = "Infrastructure Resources"
	infra.Description = "Node CPU, memory, disk and resource rankings."
	infra.Widgets = []customdashboard.Widget{
		widget("w-node-cpu", "Node CPU", "line", 0, 0, "node-cpu"),
		widget("w-node-memory", "Node Memory", "area", 1, 0, "node-memory"),
		widget("w-disk", "Disk Usage", "bar", 0, 1, "disk-usage"),
		widget("w-node-network", "Network Traffic", "line", 1, 1, "node-network-traffic"),
		widget("w-pod-status", "Pod Status Distribution", "bar", 0, 2, "pod-status-distribution"),
		widget("w-resource-top", "Resource Usage TopN", "table", 1, 2, "resource-usage-top"),
	}
	return []customdashboard.Template{
		{ID: "service-health", Name: "Service Health", Description: service.Description, Category: "service", Dashboard: service},
		{ID: "infrastructure-resources", Name: "Infrastructure Resources", Description: infra.Description, Category: "infrastructure", Dashboard: infra},
		{ID: "blank", Name: "Blank Dashboard", Description: "Empty grid with default system variables.", Category: "blank", Dashboard: blank},
	}
}

func widget(id, title, typ string, x, y int, preset string) customdashboard.Widget {
	return customdashboard.Widget{
		ID:      id,
		Title:   title,
		Type:    typ,
		Layout:  customdashboard.WidgetLayout{X: x, Y: y},
		Query:   customdashboard.WidgetQuery{Mode: customdashboard.QueryModePreset, PresetID: preset, Params: map[string]string{}},
		Display: customdashboard.WidgetDisplay{Decimals: 2},
	}
}
