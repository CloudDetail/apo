// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

type GetServiceEndpointRelationResponse struct {
	Parents       []*model.TopologyNode    `json:"parents"`        // upstream node list
	Current       *model.TopologyNode      `json:"current"`        // current service
	ChildRelation []*model.ToplogyRelation `json:"childRelations"` // downstream node call relationship list
}

type GetServiceEndpointTopologyResponse struct {
	Parents  []*model.TopologyNode `json:"parents"`  // upstream node list
	Current  *model.TopologyNode   `json:"current"`  // current service
	Children []*model.TopologyNode `json:"children"` // downstream node list
}

type GetDescendantMetricsResponse = prometheus.DescendantMetrics

type GetDescendantRelevanceResponse struct {
	ServiceName      string  `json:"serviceName"`  // service name
	EndPoint         string  `json:"endpoint"`     // Endpoint
	Group            string  `json:"group"`        // service type
	IsTraced         bool    `json:"isTraced"`     // whether to trace
	Distance         float64 `json:"distance"`     // delay curve difference
	DistanceType     string  `json:"distanceType"` // delay curve difference calculation method, there are four types of euclidean/pearson/dtw/failed/net_failed
	DelaySource      string  `json:"delaySource"`  // main source of delay unknown/self/dependency
	REDMetricsStatus string  `json:"REDStatus"`    // RED metric alarm
	LastUpdateTime   *int64  `json:"timestamp"`    // Last deployment time

	model.AlertStatus
	AlertReason model.AlertReason `json:"alertReason"`
}

type GetPolarisInferResponse = polarisanalyzer.PolarisInferRes

type GetErrorInstanceResponse struct {
	Status    string           `json:"status"`    // Bad instance status
	Instances []*ErrorInstance `json:"instances"` // error instance list
}

type ErrorInstance struct {
	Name        string            `json:"name"`        // Bad instance name
	ContainerId string            `json:"containerId"` // container ID
	NodeName    string            `json:"nodeName"`    // hostname
	Pid         int64             `json:"pid"`         // process number
	Propations  []*ErrorPropation `json:"propations"`  // error propagation info
	Logs        map[int64]float64 `json:"logs"`        // log alarm
}

type ErrorPropation struct {
	Timestamp  int64           `json:"timestamp"` // timestamp
	TraceId    string          `json:"traceId"`   // TraceId
	ErrorInfos []*ErrorInfo    `json:"errors"`    // error message
	Parents    []*InstanceNode `json:"parents"`   // upstream node list
	Current    *InstanceNode   `json:"current"`   // current node
	Children   []*InstanceNode `json:"children"`  // downstream node list
}

type ErrorInfo struct {
	Type    string `json:"type"`    // error type
	Message string `json:"message"` // error message
}

type InstanceNode struct {
	Service  string `json:"service"`
	Instance string `json:"instance"`
	IsTraced bool   `json:"isTraced"`
}

type GetLogMetricsResponse struct {
	Name        string            `json:"name"`        // Instance name
	ContainerId string            `json:"containerId"` // container ID
	NodeName    string            `json:"nodeName"`    // hostname
	Pid         int64             `json:"pid"`         // process number
	Logs        map[int64]float64 `json:"logs"`        // log alarm
	Latency     map[int64]float64 `json:"latency"`     // delay P90
	ErrorRate   map[int64]float64 `json:"errorRate"`   // error rate
}

type GetTraceMetricsResponse = GetLogMetricsResponse

type AlarmStatus struct {
	Name   string // Alarm entry
	Status bool   // alarm true: alarm false: no alarm
}

type Ratio struct {
	// Day-over-Day Growth Rate
	DayOverDay *float64 `json:"dayOverDay"`
	// Week-over-Week Growth Rate
	WeekOverDay *float64 `json:"weekOverDay"`
}

type TempChartObject struct {
	// ChartData chart data
	ChartData map[int64]float64 `json:"chartData"`
	// Value metric average
	Value *float64 `json:"value"`
	// Ratio metric Day-over-Day Growth Rate rate
	Ratio Ratio `json:"ratio"`
}

type ServiceDetail struct {
	Endpoint    string          `json:"endpoint"`
	DelaySource string          `json:"delaySource"`
	Latency     TempChartObject `json:"latency"`
	ErrorRate   TempChartObject `json:"errorRate"`
	Tps         TempChartObject `json:"tps"` // FIXME name is tps, actual requests per minute
}

type RedCharts struct {
	Latency   map[int64]float64 `json:"latency"`
	ErrorRate map[int64]float64 `json:"errorRate"`
	RPS       map[int64]float64 `json:"tps"`
}

type GetServiceREDChartsResponse map[string]map[string]RedCharts


type ServiceRes struct {
	ServiceName          string          `json:"serviceName"`
	EndpointCount        int             `json:"endpointCount"`
	ServiceDetails       []ServiceDetail `json:"serviceDetails"`
	Logs                 TempChartObject `json:"logs"`
	InfrastructureStatus string          `json:"infrastructureStatus"`
	NetStatus            string          `json:"netStatus"`
	K8sStatus            string          `json:"k8sStatus"`
	Timestamp            *int64          `json:"timestamp"`
}
type ServiceAlertRes struct {
	ServiceName string          `json:"serviceName"`
	Logs        TempChartObject `json:"logs"`

	Timestamp *int64 `json:"timestamp"`

	model.AlertStatus
	AlertReason model.AlertReason `json:"alertReason"`
}

type ServiceEndPointsRes struct {
	ServiceName    string          `json:"serviceName"`
	Namespaces     []string        `json:"namespaces"` // The namespace of the application. It may be empty
	EndpointCount  int             `json:"endpointCount"`
	ServiceDetails []ServiceDetail `json:"serviceDetails"`
}

type ServiceRYGLightRes struct {
	ServiceList []*ServiceRYGResult `json:"serviceList"`
}

type RYGStatus string

const (
	RED    RYGStatus = "red"
	YELLOW RYGStatus = "yellow"
	GREEN  RYGStatus = "green"
)

type ServiceRYGResult struct {
	ServiceName string `json:"serviceName"`

	RYGResult
}

const (
	// latency/error_rate/log_error_count/alert/replica
	MAX_RYG_SCORE = 3 * 5
)

type RYGResult struct {
	PercentScore int       `json:"percentScore"` // percentage total score
	Score        int       `json:"score"`        // total score
	Status       RYGStatus `json:"status"`

	ScoreDetail []RYGScoreDetail `json:"scoreDetail"`
}

type RYGScoreDetail struct {
	Key    string `json:"key"`
	Score  int    `json:"score"`
	Detail string `json:"detail"`
}

type InstanceData struct {
	Name      string          `json:"name"` // Instance name
	Namespace string          `json:"namespace"`
	NodeName  string          `json:"nodeName"`
	NodeIP    string          `json:"nodeIP"`
	Timestamp *int64          `json:"timestamp"`
	Latency   TempChartObject `json:"latency"`
	ErrorRate TempChartObject `json:"errorRate"`
	Tps       TempChartObject `json:"tps"`
	Logs      TempChartObject `json:"logs"`

	model.AlertStatus
	AlertReason model.AlertReason `json:"alertReason"`
}

type InstancesRes struct {
	Status string         `json:"status"`
	Data   []InstanceData `json:"data"`
}
type SetThresholdResponse struct {
}
type GetThresholdResponse struct {
	Latency   float64 `json:"latency"`
	ErrorRate float64 `json:"errorRate"`
	Tps       float64 `json:"tps"`
	Log       float64 `json:"log"`
}

type GetK8sEventsResponse struct {
	Status  string                         `json:"status"`
	Reasons []string                       `json:"reasons"`
	Data    map[string]*K8sEventStatistics `json:"data"`
}
type K8sEventStatistics struct {
	EventName   string `json:"eventName"`
	DisplayName string `json:"displayName"`
	// Normal or Warning
	Severity string              `json:"severity"`
	Counts   K8sEventCountValues `json:"counts"`
}
type K8sEventCountValues struct {
	Current   uint64 `json:"current"`
	LastWeek  uint64 `json:"lastWeek"`
	LastMonth uint64 `json:"lastMonth"`
}

func (v *K8sEventCountValues) AddCount(dao clickhouse.K8sEventsCount) {
	switch dao.TimeRange {
	case "current":
		v.Current += dao.Count
	case "lastWeek":
		v.LastWeek += dao.Count
	case "lastMonth":
		v.LastMonth += dao.Count
	}
}

type GetAlertEventsResponse struct {
	TotalCount int `json:"totalCount"`

	EventList []clickhouse.PagedAlertEvent `json:"events"`
}

type GetAlertEventsSampleResponse struct {
	EventMap map[string]map[string][]clickhouse.AlertEventSample `json:"events"`
	Status   string                                              `json:"status"`
}

type GetServiceEntryEndpointsResponse struct {
	Status string               `json:"status"`
	Data   []*EntryInstanceData `json:"data"`
}

type EntryInstanceData struct {
	ServiceName    string          `json:"serviceName"`
	Namespaces     []string        `json:"namespaces"` // The namespace of the application. It may be empty
	EndpointCount  int             `json:"endpointCount"`
	ServiceDetails []ServiceDetail `json:"serviceDetails"`

	Logs      TempChartObject `json:"logs"`
	Timestamp *int64          `json:"timestamp"`
	model.AlertStatus
	AlertReason model.AlertReason `json:"alertReason"`
}

type AlertRelatedEntry struct {
	ServiceName string   `json:"serviceName"`
	Namespaces  []string `json:"namespaces,omitempty"` // 应用所属命名空间,可能为空

	ServiceDetail

	RelatedAlertRate float64 `json:"relatedAlertRate"`
}

func (entryInstanceData *EntryInstanceData) AddNamespaces(namespaces []string) {
	if len(namespaces) == 0 {
		return
	}
	if len(entryInstanceData.Namespaces) == 0 {
		entryInstanceData.Namespaces = namespaces
	} else {
		for _, namespace := range namespaces {
			if !util.ContainsStr(entryInstanceData.Namespaces, namespace) {
				entryInstanceData.Namespaces = append(entryInstanceData.Namespaces, namespace)
			}
		}
	}
}

type GetMonitorStatusResponse struct {
	MonitorList []MonitorStatus `json:"monitorList"`
}

type MonitorStatus struct {
	MonitorName string `json:"monitorName"`
	IsAlive     bool   `json:"isAlive"`
}

type GetServiceNamespaceListResponse struct {
	NamespaceList []string `json:"namespaceList"`
}
