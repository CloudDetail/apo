// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

type GetServiceEndpointTopologyRequest struct {
	StartTime     int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime       int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Service       string `form:"service" binding:"required"`                   // query service name
	Endpoint      string `form:"endpoint" binding:"required"`                  // query Endpoint
	EntryService  string `form:"entryService"`                                 // Ingress service name
	EntryEndpoint string `form:"entryEndpoint"`                                // entry Endpoint
}

type GetServiceEndpointRelationRequest = GetServiceEndpointTopologyRequest

type GetDescendantMetricsRequest struct {
	StartTime     int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime       int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Service       string `form:"service" binding:"required"`                   // query service name
	Endpoint      string `form:"endpoint" binding:"required"`                  // query Endpoint
	Step          int64  `form:"step" binding:"min=1000000"`                   // query step size (us)
	EntryService  string `form:"entryService"`                                 // Ingress service name
	EntryEndpoint string `form:"entryEndpoint"`                                // entry Endpoint
}

type GetPolarisInferRequest struct {
	StartTime int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Step      int64  `form:"step" binding:"required"`                      // query step (us)
	Service   string `form:"service" binding:"required"`                   // query service name
	Endpoint  string `form:"endpoint" binding:"required"`                  // query Endpoint

	Lanaguage string `form:"language" json:"language"` // language of result
	Timezone  string `form:"timezone" json:"timezone"` // timezone of result
}

type GetDescendantRelevanceRequest = GetDescendantMetricsRequest

type GetErrorInstanceRequest = GetDescendantMetricsRequest

type GetErrorInstanceLogsRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Service     string `form:"service" binding:"required"`                   // query service name
	Endpoint    string `form:"endpoint" binding:"required"`                  // query Endpoint
	Instance    string `form:"instance"`                                     // instance name
	NodeName    string `form:"nodeName"`                                     // hostname
	ContainerId string `form:"containerId"`                                  // container name
	Pid         uint32 `form:"pid"`                                          // process number
}

type GetLogMetricsRequest = GetDescendantMetricsRequest
type GetLogLogsRequest = GetErrorInstanceLogsRequest

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
	StartTime   int64  `form:"startTime" binding:"required"`                 // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	ServiceName string `form:"service" binding:"required"`                   // query service name
}

type GetServiceListRequest struct {
	StartTime int64    `form:"startTime" binding:"min=0"`                    // query start time
	EndTime   int64    `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Namespace []string `form:"namespace"`
}

type GetServiceInstanceListRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	ServiceName string `form:"service" binding:"required"`                   // query service name
}

type GetServiceInstanceRequest struct {
	StartTime   int64  `form:"startTime" binding:"required"`                 // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Step        int64  `form:"step" binding:"required"`                      // step size
	ServiceName string `form:"serviceName" binding:"required"`               // application name
	Endpoint    string `form:"endpoint"`
}

type GetServiceInstanceOptionsRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	ServiceName string `form:"service" binding:"required"`                   // query service name
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
}
type GetEndPointsDataRequest struct {
	// Filter Criteria
	ServiceName  []string `form:"serviceName,omitempty"`  // application name, exact match
	Namespace    []string `form:"namespace,omitempty"`    // specify namespace, exact match
	EndpointName []string `form:"endpointName,omitempty"` // endpoint name, exact match
	GroupID      int64    `form:"groupId,omitempty"`      // Data group id

	// Query condition
	StartTime int64 `form:"startTime" binding:"required"`                 // query start time
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Step      int64 `form:"step" binding:"required"`                      // step size
	SortRule  int   `form:"sortRule" binding:"required"`                  // sort logic
}

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
	StartTime int64 `form:"startTime" binding:"required"`                 // query start time
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // query end time

	AlertFilter // filter parameters
	*PageParam  // Paging Parameters
}

type AlertFilter struct {
	Service  string   `form:"service"`
	Endpoint string   `form:"endpoint"`
	Services []string `form:"services"`

	Source   string `form:"source"`
	Group    string `form:"group"`
	Name     string `form:"name"`
	ID       string `form:"id"`
	Severity string `form:"severity"`
	Status   string `form:"status"`
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
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Service     string `form:"service" binding:"required"`                   // query service name
	Endpoint    string `form:"endpoint" binding:"required"`                  // query Endpoint
	Step        int64  `form:"step" binding:"required"`                      // query step (us)
	ShowMissTop bool   `form:"showMissTop"`                                  // whether to display the lost non-portal service
}

type GetMonitorStatusRequest struct {
	StartTime int64 `form:"startTime" binding:"min=0"`                    // query start time
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
}

type GetServiceNamespaceListRequest struct {
	StartTime int64 `form:"startTime" binding:"min=0"`
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"`
}

type GetServiceMoreUrlListRequest struct {
	StartTime   int64  `form:"startTime" binding:"required"`                 // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Step        int64  `form:"step" binding:"required"`                      // step size
	ServiceName string `form:"serviceName" binding:"required"`               // application name
	SortRule    int    `form:"sortRule" binding:"required"`                  // sort logic
}

type GetServiceREDChartsRequest struct {
	StartTime    int64    `json:"startTime"`
	EndTime      int64    `json:"endTime"`
	Step         int64    `json:"step"`
	ServiceList  []string `json:"serviceList"`
	EndpointList []string `json:"endpointList"`
}
