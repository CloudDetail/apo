// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetServiceEndpointTopologyRequest struct {
	StartTime     int64  `form:"startTime" json:"startTime" binding:"min=0"`                   // query start time
	EndTime       int64  `form:"endTime" json:"endTime"  binding:"required,gtfield=StartTime"` // query end time
	Service       string `form:"service" json:"service" binding:"required"`                    // query service name
	Endpoint      string `form:"endpoint" json:"endpoint"  binding:"required"`                 // query Endpoint
	EntryService  string `form:"entryService" json:"entryService"`                             // Ingress service name
	EntryEndpoint string `form:"entryEndpoint" json:"entryEndpoint"`                           // entry Endpoint

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetServiceEndpointRelationRequest = GetServiceEndpointTopologyRequest

type GetDescendantMetricsRequest struct {
	StartTime     int64  `form:"startTime" binding:"min=0" json:"startTime"`                  // query start time
	EndTime       int64  `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time
	Service       string `form:"service" binding:"required" json:"service"`                   // query service name
	Endpoint      string `form:"endpoint" binding:"required" json:"endpoint"`                 // query Endpoint
	Step          int64  `form:"step" binding:"min=1000000" json:"step"`                      // query step size (us)
	EntryService  string `form:"entryService" json:"entryService"`                            // Ingress service name
	EntryEndpoint string `form:"entryEndpoint" json:"entryEndpoint"`                          // entry Endpoint

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetPolarisInferRequest struct {
	StartTime int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Step      int64  `form:"step" binding:"required"`                      // query step (us)
	Service   string `form:"service" binding:"required"`                   // query service name
	Endpoint  string `form:"endpoint" binding:"required"`                  // query Endpoint

	Language string `form:"language" json:"language"` // language of result
	Timezone string `form:"timezone" json:"timezone"` // timezone of result

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetDescendantRelevanceRequest = GetDescendantMetricsRequest

type GetErrorInstanceRequest struct {
	StartTime     int64  `form:"startTime" binding:"min=0" json:"startTime"`                  // query start time
	EndTime       int64  `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time
	Service       string `form:"service" binding:"required" json:"service"`                   // query service name
	Endpoint      string `form:"endpoint" binding:"required" json:"endpoint"`                 // query Endpoint
	Step          int64  `form:"step" binding:"min=1000000" json:"step"`                      // query step size (us)
	EntryService  string `form:"entryService" json:"entryService"`                            // Ingress service name
	EntryEndpoint string `form:"entryEndpoint" json:"entryEndpoint"`                          // entry Endpoint

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetErrorInstanceLogsRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0" json:"startTime"`                  // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time
	Service     string `form:"service" binding:"required" json:"service"`                   // query service name
	Endpoint    string `form:"endpoint" binding:"required" json:"endpoint"`                 // query Endpoint
	Instance    string `form:"instance" json:"instance"`                                    // instance name
	NodeName    string `form:"nodeName" json:"nodeName"`                                    // hostname
	ContainerId string `form:"containerId" json:"containerId"`                              // container name
	Pid         uint32 `form:"pid" json:"pid"`                                              // process number

	ClusterID []string `form:"clusterIds" json:"clusterIds"`
}

type GetLogMetricsRequest = GetDescendantMetricsRequest

type GetLogLogsRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0" json:"startTime"`                  // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time
	Service     string `form:"service" binding:"required" json:"service"`                   // query service name
	Endpoint    string `form:"endpoint" binding:"required" json:"endpoint"`                 // query Endpoint
	Instance    string `form:"instance" json:"instance"`                                    // instance name
	NodeName    string `form:"nodeName" json:"nodeName"`                                    // hostname
	ContainerId string `form:"containerId" json:"containerId"`                              // container name
	Pid         uint32 `form:"pid" json:"pid"`                                              // process number

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetTraceMetricsRequest = GetDescendantMetricsRequest
type GetTraceLogsRequest = GetErrorInstanceLogsRequest

type GetThresholdRequest struct {
	ServiceName string `form:"serviceName" `
	Endpoint    string `form:"endpoint" `
	Level       string `form:"level" binding:"required"`
}

type SetThresholdRequest struct {
	ServiceName string  `form:"serviceName"`
	Endpoint    string  `form:"endpoint"`
	Level       string  `form:"level" binding:"required"`
	Latency     float64 `form:"latency" binding:"required"`
	ErrorRate   float64 `form:"errorRate" binding:"required"`
	Tps         float64 `form:"tps" binding:"required"`
	Log         float64 `form:"log" binding:"required"`
}

type GetK8sEventsRequest struct {
	StartTime   int64  `form:"startTime" binding:"required" json:"startTime"`               // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time
	ServiceName string `form:"service" binding:"required" json:"service"`                   // query service name

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetServiceListRequest struct {
	StartTime int64 `form:"startTime" binding:"min=0" json:"startTime"`                  // query start time
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
	Namespace  []string `form:"namespace" json:"namespace"`
}

type GetServiceInstanceListRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	ServiceName string `form:"service" binding:"required"`                   // query service name

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetServiceInstanceRequest struct {
	StartTime   int64  `form:"startTime" binding:"required" json:"startTime"`               // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time
	Step        int64  `form:"step" binding:"required" json:"step"`                         // step size
	ServiceName string `form:"serviceName" binding:"required" json:"serviceName"`           // application name
	Endpoint    string `form:"endpoint" json:"endpoint"`

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetServiceInstanceOptionsRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0" json:"startTime"`                  // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time
	ServiceName string `form:"service" binding:"required" json:"service"`                   // query service name

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetServiceAlertRequest struct {
	StartTime    int64    `form:"startTime" binding:"required"`                 // query start time
	EndTime      int64    `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Step         int64    `form:"step" binding:"required"`                      // step size
	ServiceNames []string `form:"serviceNames" binding:"required"`              // application name
	ReturnData   []string `form:"returnData"`
}

type GetServiceEndPointListRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	ServiceName string `form:"service"`                                      // query service name

	ClusterIDs []string `form:"clusterIds,omitempty" json:"clusterIds"`
}

type GetEndPointsDataRequest struct {
	// Filter Criteria
	ServiceName  []string `form:"serviceName,omitempty" json:"serviceName"`   // application name, exact match
	Namespace    []string `form:"namespace,omitempty" json:"namespace"`       // specify namespace, exact match
	EndpointName []string `form:"endpointName,omitempty" json:"endpointName"` // endpoint name, exact match
	GroupID      int64    `form:"groupId,omitempty" json:"groupId"`           // Data group id
	ClusterIDs   []string `form:"clusterIds,omitempty" json:"clusterIds"`     // Cluster id

	// Query condition
	StartTime int64    `form:"startTime" json:"startTime" binding:"required"`               // query start time
	EndTime   int64    `form:"endTime" json:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Step      int64    `form:"step" json:"step" binding:"required"`                         // step size
	SortRule  SortType `form:"sortRule" json:"sortRule" binding:"required"`                 // sort logic
}

type SortType int

const (
	// Sort by Day-over-Day Growth Rate Threshold
	DODThreshold SortType = iota + 1
	// Sort by mutation
	MUTATIONSORT

	SortByLatency
	SortByErrorRate
	SortByThroughput
	SortByLogErrorCount
)

type GetRygLightRequest struct {
	// Filter Criteria
	ServiceName  string `form:"serviceName"`  // application name, including matching
	EndpointName string `form:"endpointName"` // endpoint name, including matches
	Namespace    string `form:"namespace"`    // specify namespace, exact match

	// Query condition
	StartTime int64 `form:"startTime" binding:"required"`                 // query start time
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
}

type GetAlertEventsRequest struct {
	StartTime int64 `form:"startTime" binding:"required" json:"startTime"`               // query start time
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time

	AlertFilter // filter parameters
	*PageParam  // Paging Parameters
}

// AlertFilter provide params to filter alertEvents
type AlertFilter struct {
	// basic filter
	Source   string `form:"source" json:"source"`
	Group    string `form:"group" json:"group"`
	Name     string `form:"name" json:"name"`
	ID       string `form:"id" json:"id"`
	Severity string `form:"severity" json:"severity"`
	Status   string `form:"status" json:"status"`

	ClusterIDs []string `form:"clusterID" json:"clusterID"`
	Services   []string `form:"services" json:"services" binding:"required"`
	Endpoints  []string `form:"endpoints" json:"endpoints" binding:"required"`

	// Deprecated: use Services instead
	Service string `form:"service" json:"service"`
	// Deprecated: use Endpoints instead
	Endpoint string `form:"endpoint" json:"endpoint"`
}

type PageParam struct {
	CurrentPage int `form:"currentPage" json:"currentPage"`
	PageSize    int `form:"pageSize" json:"pageSize"`
}

type GetAlertEventsSampleRequest struct {
	StartTime int64 `form:"startTime" binding:"required"`                 // query start time
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // query end time

	AlertFilter // filter parameters

	SampleCount int `form:"sampleCount"` // number of samples
}

type GetServiceEntryEndpointsRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0" json:"startTime"`                  // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"` // query end time
	Service     string `form:"service" binding:"required" json:"service" `                  // query service name
	Endpoint    string `form:"endpoint" binding:"required" json:"endpoint"`                 // query Endpoint
	Step        int64  `form:"step" binding:"required" json:"step"`                         // query step (us)
	ShowMissTop bool   `form:"showMissTop" json:"showMissTop"`                              // whether to display the lost non-portal service

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetMonitorStatusRequest struct {
	StartTime int64 `form:"startTime" binding:"min=0"`                    // query start time
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
}

type GetServiceNamespaceListRequest struct {
	StartTime int64 `form:"startTime" binding:"min=0" json:"startTime"`
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime" json:"endTime"`

	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetServiceMoreUrlListRequest struct {
	StartTime   int64    `form:"startTime" json:"startTime" binding:"required"`               // query start time
	EndTime     int64    `form:"endTime" json:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Step        int64    `form:"step" json:"step" binding:"required"`                         // step size
	ServiceName string   `form:"serviceName" json:"serviceName" binding:"required"`           // application name
	SortRule    SortType `form:"sortRule" json:"sortRule" binding:"required"`                 // sort logic

	GroupID    int64    `form:"groupId" json:"groupId"`
	ClusterIDs []string `form:"clusterIds" json:"clusterIds"`
}

type GetServiceREDChartsRequest struct {
	StartTime    int64    `json:"startTime"`
	EndTime      int64    `json:"endTime"`
	Step         int64    `json:"step"`
	ServiceList  []string `json:"serviceList"`
	EndpointList []string `json:"endpointList"`
}
