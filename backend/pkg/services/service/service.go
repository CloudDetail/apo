package service

import (
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/polarisanalyzer"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

var _ Service = (*service)(nil)

type Service interface {
	// 获取上下游调用关系
	GetServiceEndpointRelation(req *request.GetServiceEndpointRelationRequest) (*response.GetServiceEndpointRelationResponse, error)
	// 获取上下游拓扑图
	GetServiceEndpointTopology(req *request.GetServiceEndpointTopologyRequest) (*response.GetServiceEndpointTopologyResponse, error)
	// 获取依赖服务的延时曲线
	GetDescendantMetrics(req *request.GetDescendantMetricsRequest) ([]response.GetDescendantMetricsResponse, error)
	// 获取依赖节点延时关联度
	GetDescendantRelevance(req *request.GetDescendantRelevanceRequest) ([]response.GetDescendantRelevanceResponse, error)
	// 获取北极星指标分析情况
	GetPolarisInfer(req *request.GetPolarisInferRequest) (*response.GetPolarisInferResponse, error)
	// 获取错误实例
	GetErrorInstance(req *request.GetErrorInstanceRequest) (*response.GetErrorInstanceResponse, error)
	// 获取错误实例故障现场日志
	GetErrorInstanceLogs(req *request.GetErrorInstanceLogsRequest) ([]clickhouse.FaultLogResult, error)
	// 获取日志相关指标
	GetLogMetrics(req *request.GetLogMetricsRequest) ([]*response.GetLogMetricsResponse, error)
	// 获取Log故障现场日志
	GetLogLogs(req *request.GetLogLogsRequest) ([]clickhouse.FaultLogResult, error)
	// 获取Trace相关指标
	GetTraceMetrics(req *request.GetTraceMetricsRequest) ([]*response.GetTraceMetricsResponse, error)
	// 获取SQL相关指标
	GetSQLMetrics(req *request.GetSQLMetricsRequest) (*response.GetSQLMetricsResponse, error)
	// 获取Trace故障现场日志
	GetTraceLogs(req *request.GetTraceLogsRequest) ([]clickhouse.FaultLogResult, error)
	// 获取服务列表
	GetServiceList(req *request.GetServiceListRequest) ([]string, error)
	// 获取服务实例列表
	// 新接口
	GetInstancesNew(startTime time.Time, endTime time.Time, step time.Duration, serviceName string, endPoint string) (res response.InstancesRes, err error)
	// 旧接口
	GetInstances(startTime time.Time, endTime time.Time, step time.Duration, serviceName string, endPoint string) (res response.InstancesRes, err error)
	// 获取服务实例列表
	// DEPRECATED
	GetServiceInstanceList(req *request.GetServiceInstanceListRequest) ([]string, error)
	// 获取服务实例下拉列表
	GetServiceInstanceOptions(req *request.GetServiceInstanceOptionsRequest) (map[string]*model.ServiceInstance, error)
	// 获取服务Endpoint列表
	GetServiceEndPointList(req *request.GetServiceEndPointListRequest) ([]string, error)
	// 获取服务入口Endpoint列表
	GetServiceEntryEndpoints(req *request.GetServiceEntryEndpointsRequest) ([]clickhouse.EntryNode, error)
	// CountK8sEvents 获取K8s事件数量
	CountK8sEvents(req *request.GetK8sEventsRequest) (*response.GetK8sEventsResponse, error)
	// GetAlertEvents 获取告警事件
	GetAlertEvents(req *request.GetAlertEventsRequest) (*response.GetAlertEventsResponse, error)

	// GetAlertEventsSample 获取告警事件
	GetAlertEventsSample(req *request.GetAlertEventsSampleRequest) (*response.GetAlertEventsSampleResponse, error)
}

type service struct {
	chRepo   clickhouse.Repo
	promRepo prometheus.Repo
	polRepo  polarisanalyzer.Repo
	dbRepo   database.Repo
}

func New(chRepo clickhouse.Repo, promRepo prometheus.Repo, polRepo polarisanalyzer.Repo, dbRepo database.Repo) Service {
	return &service{
		chRepo:   chRepo,
		promRepo: promRepo,
		polRepo:  polRepo,
		dbRepo:   dbRepo,
	}
}
