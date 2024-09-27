package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

type GetServiceEndpointRelationResponse struct {
	Parents       []clickhouse.TopologyNode    `json:"parents"`        // 上游节点列表
	Current       clickhouse.TopologyNode      `json:"current"`        // 当前服务
	ChildRelation []clickhouse.ToplogyRelation `json:"childRelations"` // 下游节点调用关系列表
}

type GetServiceEndpointTopologyResponse struct {
	Parents  []clickhouse.TopologyNode `json:"parents"`  // 上游节点列表
	Current  clickhouse.TopologyNode   `json:"current"`  // 当前服务
	Children []clickhouse.TopologyNode `json:"children"` // 下游节点列表
}

type GetDescendantMetricsResponse = prometheus.DescendantMetrics

type GetDescendantRelevanceResponse struct {
	ServiceName      string  `json:"serviceName"`  // 服务名
	EndPoint         string  `json:"endpoint"`     // Endpoint
	Distance         float64 `json:"distance"`     // 延时曲线差异
	DistanceType     string  `json:"distanceType"` // 延时曲线差异计算方式, 有euclidean/pearson/dtw/failed/net_failed四种
	DelaySource      string  `json:"delaySource"`  // 延时主要来源 unknown/self/dependency
	REDMetricsStatus string  `json:"REDStatus"`    // RED指标告警
	LastUpdateTime   *int64  `json:"timestamp"`    // 末次部署时间

	model.AlertStatus
	AlertReason model.AlertReason `json:"alertReason"`
}

type GetPolarisInferResponse = polarisanalyzer.PolarisInferRes

type GetErrorInstanceResponse struct {
	Status    string           `json:"status"`    // 错误实例状态
	Instances []*ErrorInstance `json:"instances"` // 错误实例列表
}

type ErrorInstance struct {
	Name        string            `json:"name"`        // 错误实例名
	ContainerId string            `json:"containerId"` // 容器ID
	NodeName    string            `json:"nodeName"`    // 主机名
	Pid         int64             `json:"pid"`         // 进程号
	Propations  []*ErrorPropation `json:"propations"`  // 错误传播信息
	Logs        map[int64]float64 `json:"logs"`        // 日志告警
}

type ErrorPropation struct {
	Timestamp  int64           `json:"timestamp"` // 时间戳
	TraceId    string          `json:"traceId"`   // TraceId
	ErrorInfos []*ErrorInfo    `json:"errors"`    // 错误信息
	Parents    []*InstanceNode `json:"parents"`   // 上游节点列表
	Current    *InstanceNode   `json:"current"`   // 当前节点
	Children   []*InstanceNode `json:"children"`  // 下游节点列表
}

type ErrorInfo struct {
	Type    string `json:"type"`    // 错误类型
	Message string `json:"message"` // 错误消息
}

type InstanceNode struct {
	Service  string `json:"service"`
	Instance string `json:"instance"`
	IsTraced bool   `json:"isTraced"`
}

type GetLogMetricsResponse struct {
	Name        string            `json:"name"`        // 实例名
	ContainerId string            `json:"containerId"` // 容器ID
	NodeName    string            `json:"nodeName"`    // 主机名
	Pid         int64             `json:"pid"`         // 进程号
	Logs        map[int64]float64 `json:"logs"`        // 日志告警
	Latency     map[int64]float64 `json:"latency"`     // 延时P90
	ErrorRate   map[int64]float64 `json:"errorRate"`   // 错误率
}

type GetTraceMetricsResponse = GetLogMetricsResponse

type AlarmStatus struct {
	Name   string // 告警项
	Status bool   // 是否告警 true: 告警 false: 未告警
}

type Ratio struct {
	// DayOverDay 日同比变化率
	DayOverDay *float64 `json:"dayOverDay"`
	// WeekOverDay 周同比变化率
	WeekOverDay *float64 `json:"weekOverDay"`
}

type TempChartObject struct {
	// ChartData 图表数据
	ChartData map[int64]float64 `json:"chartData"`
	// Value 指标平均值
	Value *float64 `json:"value"`
	// Ratio 指标同比变化率
	Ratio Ratio `json:"ratio"`
}

type ServiceDetail struct {
	Endpoint    string          `json:"endpoint"`
	DelaySource string          `json:"delaySource"`
	Latency     TempChartObject `json:"latency"`
	ErrorRate   TempChartObject `json:"errorRate"`
	Tps         TempChartObject `json:"tps"` // FIXME 名称为tps,实际为每分钟请求数
}

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
	Namespaces     []string        `json:"namespaces"` // 应用所属命名空间,可能为空
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
	PercentScore int       `json:"percentScore"` // 百分比总分
	Score        int       `json:"score"`        // 总分
	Status       RYGStatus `json:"status"`

	ScoreDetail []RYGScoreDetail `json:"scoreDetail"`
}

type RYGScoreDetail struct {
	Key    string `json:"key"`
	Score  int    `json:"score"`
	Detail string `json:"detail"`
}

type InstanceData struct {
	Name      string          `json:"name"` //实例名
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
	Namespaces     []string        `json:"namespaces"` // 应用所属命名空间,可能为空
	EndpointCount  int             `json:"endpointCount"`
	ServiceDetails []ServiceDetail `json:"serviceDetails"`

	Logs      TempChartObject `json:"logs"`
	Timestamp *int64          `json:"timestamp"`
	model.AlertStatus
	AlertReason model.AlertReason `json:"alertReason"`
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
