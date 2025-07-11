// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package clickhouse

import (
	"fmt"
	"log"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
)

const (
	TEMPLATE_QUERY_STORED_METRICS   = "SELECT timestamp FROM service_red_metric %s group by timestamp"
	TEMPLATE_QUERY_GROUP_RED_METRIC = "SELECT floor(toUnixTimestamp(timestamp) / %d) as time_bucket, sum(count) as total_count, sum(error_count) as total_error, sum(duration) as total_duration FROM service_red_metric %s group by time_bucket"
	TEMPLATE_QUERY_RED_METRIC_VALUE = "SELECT sum(count) as total_count, sum(error_count) as total_error, sum(duration) as total_duration FROM service_red_metric %s"
	TEMPLATE_QUERY_ENDPOINTS        = "SELECT endpoint FROM service_red_metric %s GROUP BY endpoint"
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

func (ch *chRepo) WriteServiceRedMetrics(ctx core.Context, charts *model.ServiceRedCharts) error {
	for endpoint, redValues := range charts.EndPointCharts {
		if err := ch.writeEndpointRedCharts(ctx, charts.Service, endpoint, redValues); err != nil {
			return err
		}
	}
	return nil
}

func (ch *chRepo) writeEndpointRedCharts(ctx core.Context, service *model.Service, endpoint string, redValues map[int64]*model.RedMetricValue) error {
	metrics, err := ch.queryToSendRedMetrics(ctx, service, endpoint, redValues)
	if err != nil {
		return err
	}
	if len(metrics) == 0 {
		return nil
	}

	batch, err := ch.GetContextDB(ctx).PrepareBatch(ctx.GetContext(), `
		INSERT INTO service_red_metric (timestamp, cluster_id, source, service_id, service_name, endpoint, count, error_count, duration)
		VALUES
	`)

	if err != nil {
		return err
	}
	for _, metric := range metrics {
		if err := batch.Append(
			time.Unix(metric.Timestamp, 0).UTC(),
			service.ClusterId,
			service.Source,
			service.Id,
			service.Name,
			endpoint,
			metric.Count,
			metric.ErrorCount,
			metric.Duration); err != nil {

			log.Println("Failed to send data:", err)
			continue
		}
	}
	if err := batch.Send(); err != nil {
		return err
	}
	return nil
}

func (ch *chRepo) queryToSendRedMetrics(ctx core.Context, service *model.Service, endpoint string, redValues map[int64]*model.RedMetricValue) ([]*ServiceRedMetric, error) {
	result := make([]*ServiceRedMetric, 0)
	if (len(redValues)) == 0 {
		return result, nil
	}
	var (
		minTs int64
		maxTs int64
	)
	for ts := range redValues {
		if minTs == 0 {
			minTs = ts
		} else if minTs > ts {
			minTs = ts
		}
		if maxTs == 0 {
			maxTs = ts
		} else if maxTs < ts {
			maxTs = ts
		}
	}
	queryBuilder := NewQueryBuilder().
		Between("timestamp", minTs/1000000, maxTs/1000000).
		Equals("cluster_id", service.ClusterId).
		Equals("source", service.Source).
		Equals("service_id", service.Id).
		Equals("service_name", service.Name).
		Equals("endpoint", endpoint)
	query := fmt.Sprintf(TEMPLATE_QUERY_STORED_METRICS, queryBuilder.String())

	// Query list data
	metrics := []StoredMetric{}
	err := ch.GetContextDB(ctx).Select(ctx.GetContext(), &metrics, query, queryBuilder.values...)
	if err != nil {
		return nil, err
	}
	points := make(map[int64]bool)
	for _, metric := range metrics {
		points[metric.Timestamp.Unix()] = true
	}
	toWriteMetrics := make([]*ServiceRedMetric, 0)
	for timestamp, queriedMetrics := range redValues {
		if _, found := points[timestamp/1000000]; !found {
			toWriteMetrics = append(toWriteMetrics, &ServiceRedMetric{
				Timestamp:  timestamp / 1000000,
				Count:      uint32(queriedMetrics.Count),
				ErrorCount: uint32(queriedMetrics.ErrorCount),
				Duration:   uint64(queriedMetrics.Duration),
			})
		}
	}
	return toWriteMetrics, nil
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

type Endpoint struct {
	Endpoint string `ch:"endpoint"`
}

type StoredMetric struct {
	Timestamp time.Time `ch:"timestamp"`
}
