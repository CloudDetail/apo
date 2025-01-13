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

type GetServiceEndpointRelationRequest struct {
	GetServiceEndpointTopologyRequest

	RemoveClientCall  bool `form:"removeClientCall"`  // 是否移除对外调用关系
	WithTopologyLevel bool `form:"withTopologyLevel"` // 是否携带拓扑层级
}

type GetDescendantMetricsRequest struct {
	StartTime     int64  `form:"startTime" binding:"min=0"`                    // query start time
	EndTime       int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Service       string `form:"service" binding:"required"`                   // query service name
	Endpoint      string `form:"endpoint" binding:"required"`                  // query Endpoint
	Step          int64  `form:"step" binding:"min=1000000"`                   // query step size (us)
	EntryService  string `form:"entryService"`                                 // Ingress service name
	EntryEndpoint string `form:"entryEndpoint"`                                // entry Endpoint
}

type GetDescendantAnormalEventRequest struct {
	StartTime     int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime       int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Service       string `form:"service" binding:"required"`                   // 查询服务名
	Endpoint      string `form:"endpoint" binding:"required"`                  // 查询Endpoint
	EntryService  string `form:"entryService"`                                 // 入口服务名
	EntryEndpoint string `form:"entryEndpoint"`                                // 入口Endpoint

	*PageParam // 分页参数
}

type GetDescendantAlertContributaionRequest struct {
	GetDescendantMetricsRequest

	// 贡献度排序方式
	SortBy string `form:"sortBy"`
}

type GetDescendantAnormalTrendRequest struct {
	StartTime int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Service   string `form:"service" binding:"required"`                   // 查询服务名
	Endpoint  string `form:"endpoint" binding:"required"`                  // 查询Endpoint
	Step      int64  `form:"step" binding:"min=1000000"`                   // 查询步长(us)

	AnormalTypes string `form:"anormalTypes" json:"anormalTypes"` //要查询的异常类型
}

type GetDescendantAnormalDeltaEventRequest struct {
	StartTime int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Service   string `form:"service" binding:"required"`                   // 查询服务名
	Endpoint  string `form:"endpoint" binding:"required"`                  // 查询Endpoint
	Step      int64  `form:"step" binding:"min=1000000"`                   // 查询步长(us)

	DeltaStartTime int64 `form:"deltaStartTime" binding:"min=0"` // 播放开始时间
	DeltaEndTime   int64 `form:"deltaEndTime" binding:"min=0"`   // 播放结束时间

	AnormalTypes string `form:"anormalTypes" json:"anormalTypes"` //要查询的异常类型
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

type GetDescendantRelevanceRequest struct {
	GetDescendantMetricsRequest

	SortBy    string `form:"sortBy"` // latency, errorRate
	WithAlert string `form:"withAlert"`
}

type GetErrorInstanceRequest struct {
	GetDescendantMetricsRequest

	IgnoreEmptyPropationsInstance bool `form:"ignoreEmptyPropationsInstance"`
}

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
	StartTime int64  `form:"startTime" binding:"required"`                 // query start time
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	SortBy    string `form:"sortBy"`                                       // 排序参数

	AlertFilter // filter parameters
	*PageParam  // Paging Parameters
}

type AlertFilter struct {
	Service  string `form:"service"`
	Endpoint string `form:"endpoint"`

	Source   string `form:"source"`
	Group    string `form:"group"`
	Name     string `form:"name"`
	ID       string `form:"id"`
	Severity string `form:"severity"`
	Status   string `form:"status"`

	WithMutation bool `form:"withMutation"`
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

type GetAnomalySpanRequest struct {
	StartTime  int64  `json:"startTime" binding:"required"`
	EndTime    int64  `json:"endTime" binding:"required"`
	Service    string `json:"service" binding:"required"`
	ContentKey string `json:"contentKey" binding:"required"`
	Reason     string `json:"reason"`
	IsSlow     string `json:"isSlow"`
	IsError    string `json:"isError"`
	*PageParam `json:",inline"`
}

type GetServiceMoreUrlListRequest struct {
	StartTime   int64  `form:"startTime" binding:"required"`                 // query start time
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // query end time
	Step        int64  `form:"step" binding:"required"`                      // step size
	ServiceName string `form:"serviceName" binding:"required"`               // application name
	SortRule    int    `form:"sortRule" binding:"required"`                  // sort logic
}
