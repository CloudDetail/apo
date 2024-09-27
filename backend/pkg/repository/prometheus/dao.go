package prometheus

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Repo interface {
	// ========== span_trace_duration_bucket Start ==========
	// 基于服务列表、URL列表和时段、步长，查询P90曲线
	QueryRangePercentile(startTime int64, endTime int64, step int64, services []string, endpoints []string) ([]DescendantMetrics, error)
	// 查询实例的P90延时曲线
	QueryInstanceP90(startTime int64, endTime int64, step int64, endpoint string, instance *model.ServiceInstance) (map[int64]float64, error)
	// ========== span_trace_duration_bucket END ==========

	// ========== span_trace_duration_count Start ==========
	// 查询服务列表
	GetServiceList(startTime int64, endTime int64) ([]string, error)
	// 查询服务实例列表, URL允许为空
	GetInstanceList(startTime int64, endTime int64, serviceName string, url string) (*model.ServiceInstances, error)
	// 查询活跃实例列表
	GetActiveInstanceList(startTime int64, endTime int64, serviceName string) (*model.ServiceInstances, error)
	// 查询服务Endpoint列表，服务允许为空
	GetServiceEndPointList(startTime int64, endTime int64, serviceName string) ([]string, error)
	GetMultiServicesInstanceList(startTime int64, endTime int64, services []string) (map[string]*model.ServiceInstances, error)
	// 查询服务实例失败率
	QueryInstanceErrorRate(startTime int64, endTime int64, step int64, endpoint string, instance *model.ServiceInstance) (map[int64]float64, error)
	// ========== span_trace_duration_count END ==========

	QueryData(searchTime time.Time, query string) ([]MetricResult, error)
	QueryRangeData(startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error)
	QueryLatencyData(searchTime time.Time, query string) ([]MetricResult, error)
	QueryRangeLatencyData(startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error)
	QueryErrorRateData(searchTime time.Time, query string) ([]MetricResult, error)
	QueryRangeErrorData(startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error)

	// ========== originx_logparser_level_count_total Start ==========
	// 查询实例日志Error数
	QueryLogCountByInstanceId(instance *model.ServiceInstance, startTime int64, endTime int64, step int64) (map[int64]float64, error)
	// ========== originx_logparser_level_count_total END ==========

	QueryAggMetricsWithFilter(pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, granularity Granularity, filterKVs ...string) ([]MetricResult, error)
	QueryRangeAggMetricsWithFilter(pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, step int64, granularity Granularity, filterKVs ...string) ([]MetricResult, error)

	// originx_process_start_time
	QueryProcessStartTime(startTime time.Time, endTime time.Time, instances []*model.ServiceInstance) (map[model.ServiceInstance]int64, error)
	GetApi() v1.API
	GetRange() string
}

type promRepo struct {
	api       v1.API
	promRange string
}

func New(
	logger *zap.Logger,
	address string,
	storage string) (Repo, error) {
	promRange := "le"
	if storage == config.PROM_STORAGE_VM {
		promRange = "vmrange"
	}

	prometheusClient, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to promethues: %s", err)
	}
	api := v1.NewAPI(prometheusClient)
	// Debug 日志等级时使用包装的Conn，输出执行SQL的耗时
	if logger.Level() == zap.DebugLevel {
		return &promRepo{
			api: &WrappedApi{
				API:    api,
				logger: logger,
			},
			promRange: promRange,
		}, nil
	} else {
		return &promRepo{
			api:       api,
			promRange: promRange,
		}, nil
	}
}

func (repo *promRepo) GetApi() v1.API {
	return repo.api
}

func (repo *promRepo) GetRange() string {
	return repo.promRange
}

type Labels struct {
	ContainerID string `json:"container_id"`
	ContentKey  string `json:"content_key"`
	Instance    string `json:"instance"`
	IsError     string `json:"is_error"`
	Job         string `json:"job"`
	NodeName    string `json:"node_name"`
	POD         string `json:"pod"`
	SvcName     string `json:"svc_name"`
	TopSpan     string `json:"top_span"`
	PID         string `json:"pid"`
	PodName     string `json:"pod_name"`
	Namespace   string `json:"namespace"`
	NodeIP      string `json:"node_ip"`

	DBSystem string `json:"db_system"`
	DBName   string `json:"db_name"`
	// Name, currently represents the Opertaion section in SQL
	// e.g: SELECT trip
	Name  string `json:"name"`
	DBUrl string `json:"db_url"`
}

type MetricResult struct {
	Metric Labels   `json:"metric"`
	Values []Points `json:"values"`
}

type Points struct {
	TimeStamp int64
	Value     float64
}
