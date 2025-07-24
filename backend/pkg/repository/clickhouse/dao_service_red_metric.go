// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	TEMPLATE_QUERY_GROUP_RED_METRIC = "SELECT floor(toUnixTimestamp(timestamp) / %d) as time_bucket, sum(count) as total_count, sum(error_count) as total_error, sum(duration) as total_duration FROM service_red_metric %s group by time_bucket"
	TEMPLATE_QUERY_RED_METRIC_VALUE = "SELECT sum(count) as total_count, sum(error_count) as total_error, sum(duration) as total_duration FROM service_red_metric %s"
	TEMPLATE_QUERY_SERVICES        = "SELECT cluster_id, source, service_name, service_id FROM service_red_metric %s GROUP BY cluster_id, source, service_name, service_id"
	TEMPLATE_QUERY_ENDPOINTS       = "SELECT endpoint FROM service_red_metric %s GROUP BY endpoint"
)

func (ch *chRepo) QueryGroupServiceRedMetrics(ctx core.Context, startTime int64, endTime int64, clusterId string, serviceName string, endpoint string, step int64) ([]BucketRedMetric, error) {
	queryBuilder := NewQueryBuilder().
		Between("toUnixTimestamp(timestamp)", startTime/1000000, endTime/1000000).
		EqualsNotEmpty("cluster_id", clusterId).
		EqualsNotEmpty("service_name", serviceName).
		EqualsNotEmpty("endpoint", endpoint)
	query := fmt.Sprintf(TEMPLATE_QUERY_GROUP_RED_METRIC, step/1000000, queryBuilder.String())
	result := []BucketRedMetric{}
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &result, query, queryBuilder.values...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (ch *chRepo) QueryGroupServiceRedMetricValue(ctx core.Context, startTime int64, endTime int64, clusterId string, serviceName string, endpoint string) (*GroupRedMetric, error) {
	queryBuilder := NewQueryBuilder().
		Between("toUnixTimestamp(timestamp)", startTime/1000000, endTime/1000000).
		EqualsNotEmpty("cluster_id", clusterId).
		EqualsNotEmpty("service_name", serviceName).
		EqualsNotEmpty("endpoint", endpoint)
	query := fmt.Sprintf(TEMPLATE_QUERY_RED_METRIC_VALUE, queryBuilder.String())
	metrics := []GroupRedMetric{}
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &metrics, query, queryBuilder.values...)
	if err != nil {
		return nil, err
	}
	if len(metrics) != 1 {
		return &GroupRedMetric{
			Duration: endTime - startTime,
		}, nil
	}
	metric := &metrics[0]
	metric.Duration = endTime - startTime
	return metric, nil
}

func (ch *chRepo) QueryServices(ctx core.Context, startTime int64, endTime int64, clusterId string) ([]*model.Service, error) {
	queryBuilder := NewQueryBuilder().
		Between("toUnixTimestamp(timestamp)", startTime/1000000, endTime/1000000).
		EqualsNotEmpty("cluster_id", clusterId)
		query := fmt.Sprintf(TEMPLATE_QUERY_SERVICES, queryBuilder.String())
	services := []GroupService{}
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &services, query, queryBuilder.values...)
	if err != nil {
		return nil, err
	}
	result := make([]*model.Service, 0)
	for _, service := range services {
		result = append(result, &model.Service{
			ClusterId: service.ClusterId,
			Source:    service.Source,
			Id:        service.ServiceId,
			Name:      service.ServiceName,
		})
	}
	return result, nil
}

func (ch *chRepo) QueryServiceEndpoints(ctx core.Context, startTime int64, endTime int64, clusterId string, serviceName string) ([]string, error) {
	queryBuilder := NewQueryBuilder().
		Between("toUnixTimestamp(timestamp)", startTime/1000000, endTime/1000000).
		EqualsNotEmpty("cluster_id", clusterId).
		EqualsNotEmpty("service_name", serviceName)
	query := fmt.Sprintf(TEMPLATE_QUERY_ENDPOINTS, queryBuilder.String())
	endpoints := []Endpoint{}
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &endpoints, query, queryBuilder.values...)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	for _, endpoint := range endpoints {
		result = append(result, endpoint.Endpoint)
	}
	return result, nil
}

type ServiceRedMetric struct {
	Timestamp  int64  `ch:"timestamp" json:"timestamp"`
	Count      uint32 `ch:"count" json:"count"`
	ErrorCount uint32 `ch:"error_count" json:"errorCount"`
	Duration   uint64 `ch:"duration" json:"duration"`
}

type BucketRedMetric struct {
	TimeBucket    float64 `ch:"time_bucket"`
	TotalCount    uint64  `ch:"total_count"`
	TotalError    uint64  `ch:"total_error"`
	TotalDuration uint64  `ch:"total_duration"`
}

type GroupRedMetric struct {
	Duration      int64
	TotalCount    uint64 `ch:"total_count"`
	TotalError    uint64 `ch:"total_error"`
	TotalDuration uint64 `ch:"total_duration"`
}

func (metric *GroupRedMetric) GetAvgLatency() float64 {
	if metric.TotalCount == 0 {
		return 0.0
	}
	return float64(metric.TotalDuration) / float64(metric.TotalCount)
}

func (metric *GroupRedMetric) GetErrorRate() float64 {
	if metric.TotalCount == 0 {
		return 0.0
	}
	return 100.0 * float64(metric.TotalError) / float64(metric.TotalCount)
}

func (metric *GroupRedMetric) GetTpm() float64 {
	if metric.TotalCount == 0 {
		return 0.0
	}
	return 60_000_000.0 * float64(metric.TotalCount) / float64(metric.Duration)
}

type GroupService struct {
	ClusterId   string `ch:"cluster_id"`
	Source      string `ch:"source"`
	ServiceId   string `ch:"service_id"`
	ServiceName string `ch:"service_name"`
}

type Endpoint struct {
	Endpoint string `ch:"endpoint"`
}
