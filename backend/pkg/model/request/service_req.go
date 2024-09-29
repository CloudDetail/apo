package request

type GetServiceEndpointTopologyRequest struct {
	StartTime     int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime       int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Service       string `form:"service" binding:"required"`                   // 查询服务名
	Endpoint      string `form:"endpoint" binding:"required"`                  // 查询Endpoint
	EntryService  string `form:"entryService"`                                 // 入口服务名
	EntryEndpoint string `form:"entryEndpoint"`                                // 入口Endpoint
}

type GetServiceEndpointRelationRequest = GetServiceEndpointTopologyRequest

type GetDescendantMetricsRequest struct {
	StartTime     int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime       int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Service       string `form:"service" binding:"required"`                   // 查询服务名
	Endpoint      string `form:"endpoint" binding:"required"`                  // 查询Endpoint
	Step          int64  `form:"step" binding:"min=1000000"`                   // 查询步长(us)
	EntryService  string `form:"entryService"`                                 // 入口服务名
	EntryEndpoint string `form:"entryEndpoint"`                                // 入口Endpoint
}

type GetPolarisInferRequest struct {
	StartTime int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Step      int64  `form:"step" binding:"required"`                      // 查询步长(us)
	Service   string `form:"service" binding:"required"`                   // 查询服务名
	Endpoint  string `form:"endpoint" binding:"required"`                  // 查询Endpoint
}

type GetDescendantRelevanceRequest = GetDescendantMetricsRequest

type GetErrorInstanceRequest = GetDescendantMetricsRequest

type GetErrorInstanceLogsRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Service     string `form:"service" binding:"required"`                   // 查询服务名
	Endpoint    string `form:"endpoint" binding:"required"`                  // 查询Endpoint
	Instance    string `form:"instance"`                                     // 实例名
	NodeName    string `form:"nodeName"`                                     // 主机名
	ContainerId string `form:"containerId"`                                  // 容器名
	Pid         uint32 `form:"pid"`                                          // 进程号
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
	StartTime   int64  `form:"startTime" binding:"required"`                 // 查询开始时间
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	ServiceName string `form:"service" binding:"required"`                   // 查询服务名
}

type GetServiceListRequest struct {
	StartTime int64 `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
}

type GetServiceInstanceListRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	ServiceName string `form:"service" binding:"required"`                   // 查询服务名
}

type GetServiceInstanceOptionsRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	ServiceName string `form:"service" binding:"required"`                   // 查询服务名
}

type GetServiceAlertRequest struct {
	StartTime    int64    `form:"startTime" binding:"required"`                 // 查询开始时间
	EndTime      int64    `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Step         int64    `form:"step" binding:"required"`                      // 步长
	ServiceNames []string `form:"serviceNames" binding:"required"`              // 应用名
	ReturnData   []string `form:"returnData"`
}

type GetServiceEndPointListRequest struct {
	StartTime   int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime     int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	ServiceName string `form:"service"`                                      // 查询服务名
}
type GetEndPointsDataRequest struct {
	// 筛选条件
	ServiceName  string `form:"serviceName"`  // 应用名,包含匹配
	EndpointName string `form:"endpointName"` // 端点名,包含匹配
	Namespace    string `form:"namespace"`    // 指定命名空间,完全匹配

	// 查询条件
	StartTime int64 `form:"startTime" binding:"required"`                 // 查询开始时间
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Step      int64 `form:"step" binding:"required"`                      // 步长
	SortRule  int   `form:"sortRule" binding:"required"`                  //排序逻辑
}

type GetRygLightRequest struct {
	// 筛选条件
	ServiceName  string `form:"serviceName"`  // 应用名,包含匹配
	EndpointName string `form:"endpointName"` // 端点名,包含匹配
	Namespace    string `form:"namespace"`    // 指定命名空间,完全匹配

	// 查询条件
	StartTime int64 `form:"startTime" binding:"required"`                 // 查询开始时间
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
}

type GetAlertEventsRequest struct {
	StartTime int64 `form:"startTime" binding:"required"`                 // 查询开始时间
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间

	AlertFilter // 过滤参数
	*PageParam  // 分页参数
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
}

type PageParam struct {
	CurrentPage int `form:"currentPage" json:"currentPage"`
	PageSize    int `form:"pageSize" json:"pageSize"`
}

type GetAlertEventsSampleRequest struct {
	StartTime int64 `form:"startTime" binding:"required"`                 // 查询开始时间
	EndTime   int64 `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间

	AlertFilter // 过滤参数

	SampleCount int `form:"sampleCount"` // 采样数量
}

type GetServiceEntryEndpointsRequest struct {
	StartTime int64  `form:"startTime" binding:"min=0"`                    // 查询开始时间
	EndTime   int64  `form:"endTime" binding:"required,gtfield=StartTime"` // 查询结束时间
	Service   string `form:"service" binding:"required"`                   // 查询服务名
	Endpoint  string `form:"endpoint" binding:"required"`                  // 查询Endpoint
	Step      int64  `form:"step" binding:"required"`                      // 查询步长(us)
}
