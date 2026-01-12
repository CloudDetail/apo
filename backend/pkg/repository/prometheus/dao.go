// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/model"
	pmodel "github.com/prometheus/common/model"
	prommodel "github.com/prometheus/common/model"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Repo interface {
	// ========== span_trace_duration_bucket Start ==========
	// Query the P90 curve based on the service list, URL list, time period and step size.
	QueryRangePercentile(ctx core.Context, startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error)
	// Query the P90 delay curve of the instance
	QueryInstanceP90(ctx core.Context, startTime int64, endTime int64, step int64, endpoint string, instance *model.ServiceInstance) (map[int64]float64, error)
	// ========== span_trace_duration_bucket END ==========

	// ========== span_trace_duration_count Start ==========
	// Query the service list
	GetServiceList(ctx core.Context, startTime int64, endTime int64, filter PQLFilter) ([]string, error)
	// 根据过滤规则查询服务列表
	GetServiceListByFilter(ctx core.Context, startTime time.Time, endTime time.Time, filterKVs ...string) ([]string, error)
	// 基于DatabaseURL,IP,Port查询上游服务列表
	GetServiceListByDatabase(ctx core.Context, startTime, endTime time.Time, dbURL, dbIP, dbPort string) ([]string, error)
	// Query the service instance list. The URL can be empty.
	GetServiceWithNamespace(ctx core.Context, startTime, endTime int64, namespace []string) (map[string][]string, error)
	// GetServiceNamespace  Get service's namespaces.
	GetServiceNamespace(ctx core.Context, startTime, endTime int64, service string) ([]string, error)
	// GetInstanceList query service instance list. URL can be empty.
	GetInstanceList(ctx core.Context, startTime int64, endTime int64, serviceName string, url string) (*model.ServiceInstances, error)
	// Query the list of active instances
	GetActiveInstanceList(ctx core.Context, startTime int64, endTime int64, clusterId string, serviceNames []string) (*model.ServiceInstances, error)
	// Query the service Endpoint list. The service permission is empty.
	GetServiceEndPointList(ctx core.Context, startTime int64, endTime int64, serviceName string) ([]string, error)
	// Query the service Endpoint list. The service permission is empty.
	GetServiceEndPointListByPQLFilter(ctx core.Context, startTime int64, endTime int64, filter PQLFilter) ([]string, error)
	// Query service instance failure rate
	QueryInstanceErrorRate(ctx core.Context, startTime int64, endTime int64, step int64, endpoint string, instance *model.ServiceInstance) (map[int64]float64, error)
	// ========== span_trace_duration_count END ==========

	QueryData(ctx core.Context, searchTime time.Time, query string) ([]MetricResult, error)
	QueryRangeData(ctx core.Context, startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error)
	QueryLatencyData(ctx core.Context, searchTime time.Time, query string) ([]MetricResult, error)
	QueryRangeLatencyData(ctx core.Context, startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error)
	QueryErrorRateData(ctx core.Context, searchTime time.Time, query string) ([]MetricResult, error)
	QueryRangeErrorData(ctx core.Context, startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error)

	// ========== originx_logparser_level_count_total Start ==========
	// Query the number of errors in the instance log
	QueryLogCountByInstanceId(ctx core.Context, instance *model.ServiceInstance, startTime int64, endTime int64, step int64) (map[int64]float64, error)
	// QueryInstanceLogRangeData query instance-level log graphs
	QueryInstanceLogRangeData(ctx core.Context, pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, stepMicroS int64, granularity Granularity, podFilterKVs, vmFilterKVs []string) ([]MetricResult, error)
	// ========== originx_logparser_level_count_total END ==========

	// ========== db_duration_bucket Start ==========
	// Query the P90 curve based on the service list, URL list, time period and step size.
	QueryDbRangePercentile(ctx core.Context, startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error)
	// ========== db_duration_bucket END ==========

	// ========== external_duration_bucket Start ==========
	// Query the P90 curve based on the service list, URL list, time period and step size.
	QueryExternalRangePercentile(ctx core.Context, startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error)
	// ========== external_duration_bucket END ==========

	// ========== mq_duration_bucket Start ==========
	// Query the P90 curve based on the service list, URL list, time period and step size.
	QueryMqRangePercentile(ctx core.Context, startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error)
	// ========== mq_duration_nanoseconds_bucket END ==========

	QueryAggMetricsWithFilter(ctx core.Context, pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, granularity Granularity, filterKVs ...string) ([]MetricResult, error)
	QueryRangeAggMetricsWithFilter(ctx core.Context, pqlTemplate AggPQLWithFilters, startTime int64, endTime int64, step int64, granularity Granularity, filterKVs ...string) ([]MetricResult, error)
	// originx_process_start_time
	QueryProcessStartTime(ctx core.Context, startTime time.Time, endTime time.Time, instances []*model.ServiceInstance) (map[model.ServiceInstance]int64, error)
	GetApi() v1.API
	GetRange() string

	LabelValues(ctx core.Context, expr string, label string, startTime, endTime int64) (prommodel.LabelValues, error)
	QueryResult(ctx core.Context, expr string, regex string, startTime, endTime int64) ([]string, error)

	GetNamespaceList(ctx core.Context, startTime int64, endTime int64, filter PQLFilter) ([]string, error)
	GetNamespaceWithService(ctx core.Context, startTime, endTime int64) (map[string][]string, error)

	GetPodList(ctx core.Context, startTime int64, endTime int64, nodeName string, namespace string, podName string) ([]*model.Pod, error)
	QueryWithPQLFilter

	QueryRangeWithP9xBuilder(ctx core.Context, builder *UnionP9xBuilder, tRange v1.Range) (pmodel.Value, v1.Warnings, error)
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
	// Use the wrapped Conn at the Debug log level, and output the time taken to execute SQL.
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
	Namespace   string `json:"namespace"`
	ClusterID   string `json:"cluster_id"`
	NodeIP      string `json:"node_ip"`

	DBSystem string `json:"db_system"`
	DBName   string `json:"db_name"`
	// Name, currently represents the Opertaion section in SQL
	// e.g: SELECT trip
	Name     string `json:"name"`
	DBUrl    string `json:"db_url"`
	PeerIP   string `json:"peer_ip"`
	PeerPort string `json:"peer_port"`

	MonitorName string `json:"monitor_name"`
}

func (l *Labels) ExtractGran(granularity []string, metric prommodel.LabelSet) {
	for _, name := range granularity {
		val, find := metric[prommodel.LabelName(name)]
		if !find {
			continue
		}
		l.SetValue(name, string(val))
	}
}

// Extract extract the required label
// Changes of Labels field need to be synchronized
func (l *Labels) Extract(metric prommodel.Metric) {
	for name, value := range metric {
		l.SetValue(string(name), string(value))
	}
}

func (l *Labels) SetValue(name string, value string) {
	switch name {
	case "container_id":
		l.ContainerID = string(value)
	case "content_key":
		l.ContentKey = string(value)
	case "instance":
		l.Instance = string(value)
	case "is_error":
		l.IsError = string(value)
	case "job":
		l.Job = string(value)
	case "node_name":
		l.NodeName = string(value)
	case "pod":
		l.POD = string(value)
	case "svc_name":
		l.SvcName = string(value)
	case "top_span":
		l.TopSpan = string(value)
	case "pid":
		l.PID = string(value)
	case "namespace":
		l.Namespace = string(value)
	case "cluster_id":
		l.ClusterID = string(value)
	case "db_system":
		l.DBSystem = string(value)
	case "db_name":
		l.DBName = string(value)
	case "name":
		l.Name = string(value)
	case "db_url":
		l.DBUrl = string(value)
	case "monitor_name":
		l.MonitorName = string(value)
	case "node_ip":
		l.NodeIP = string(value)
	case "host_ip":
		l.NodeIP = string(value)
	case "host_name":
		l.NodeName = string(value)
	case "pod_name":
		l.POD = string(value)
	case "peer_ip":
		l.PeerIP = string(value)
	case "peer_port":
		l.PeerPort = string(value)
	}
}

type MetricResult struct {
	Metric Labels   `json:"metric"`
	Values []Points `json:"values"`
}

type Points struct {
	TimeStamp int64
	Value     float64
}
