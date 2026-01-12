// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheus_model "github.com/prometheus/common/model"
)

const (
	TIME_SECOND = int64(time.Second)
	TIME_MINUTE = 60 * TIME_SECOND
	TIME_HOUR   = 60 * TIME_MINUTE
	TIME_DAY    = TIME_HOUR * 24
)

// Query the P90 curve based on the service list, URL list, time period and step size.
func (repo *promRepo) QueryRangePercentile(ctx core.Context, startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error) {
	svcs, endpoints, _ := nodes.GetLabels(model.GROUP_SERVICE)
	if len(svcs) == 0 {
		return nil, nil
	}

	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}
	qb := getSpanTraceP9xSql(repo.GetRange(), tRange.Step, svcs, endpoints)
	res, _, err := repo.QueryRangeWithP9xBuilder(ctx, qb, tRange)
	if err != nil {
		return nil, err
	}

	return getDescendantMetrics("svc_name", "content_key", tRange, res), nil
}

func getSpanTraceP9xSql(promRange string, step time.Duration, svcs []string, endpoints []string) *UnionP9xBuilder {
	builder := NewUnionP9xBuilder(
		"0.9",
		"kindling_span_trace_duration_nanoseconds_bucket",
		[]string{promRange, "content_key", "svc_name"},
		step,
	)
	builder.AddCondition("svc_name", svcs)
	builder.AddCondition("content_key", endpoints)
	return builder
}

func (repo *promRepo) QueryInstanceP90(ctx core.Context, startTime int64, endTime int64, step int64, endpoint string, instance *model.ServiceInstance) (map[int64]float64, error) {
	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}

	extraCondition := ""
	if instance.PodName != "" {
		extraCondition = fmt.Sprintf("pod='%s'", instance.PodName)
	} else if instance.ContainerId != "" {
		extraCondition = fmt.Sprintf("node_name='%s', container_id='%s'", instance.NodeName, instance.ContainerId)
	} else {
		// VM scenario
		extraCondition = fmt.Sprintf("node_name='%s', pid='%d'", instance.NodeName, instance.Pid)
	}
	qb := getSpanTraceInstanceP9xSql(repo.GetRange(), tRange.Step, endpoint, extraCondition)

	res, _, err := repo.QueryRangeWithP9xBuilder(ctx, qb, tRange)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]float64, 0)
	values, ok := res.(prometheus_model.Matrix)
	if !ok {
		return result, nil
	}
	if len(values) == 1 {
		val := values[0]
		for _, pair := range val.Values {
			result[int64(pair.Timestamp)*1000] = float64(pair.Value / 1000) // return time is us
		}
	}
	return result, nil
}

func getSpanTraceInstanceP9xSql(promRange string, step time.Duration, endpoint string, extraCondition string) *UnionP9xBuilder {
	builder := NewUnionP9xBuilder(
		"0.9",
		"kindling_span_trace_duration_nanoseconds_bucket",
		[]string{promRange},
		step,
	)
	builder.AddCondition("content_key", []string{endpoint})
	builder.AddExtraCondition(extraCondition)

	return builder
}

type DescendantMetrics struct {
	ServiceName string         `json:"serviceName"` // service name
	EndPoint    string         `json:"endpoint"`    // Endpoint
	LatencyP90  []MetricsPoint `json:"latencyP90"`  // P90 curve value
}

type MetricsPoint struct {
	Timestamp int64   `json:"timestamp"` // time (microseconds)
	Value     float64 `json:"value"`     // value
}

func getDescendantMetrics(svcNameLabel prometheus_model.LabelName, contentKeyLabel prometheus_model.LabelName, tRange v1.Range, res prometheus_model.Value) []DescendantMetrics {
	result := make([]DescendantMetrics, 0)
	values, ok := res.(prometheus_model.Matrix)
	if !ok {
		return result
	}

	for _, val := range values {
		svcName, ok := val.Metric[svcNameLabel]
		if !ok {
			continue
		}
		contentKey, ok := val.Metric[contentKeyLabel]
		if !ok {
			continue
		}

		ts := DescendantMetrics{
			ServiceName: string(svcName),
			EndPoint:    string(contentKey),
			LatencyP90:  []MetricsPoint{},
		}

		tsMark := tRange.Start.UnixMicro()
		for _, pair := range val.Values {
			// The time point at which the data is not obtained is filled with 0
			for tsMark < int64(pair.Timestamp)*1000 {
				ts.LatencyP90 = append(ts.LatencyP90, MetricsPoint{
					Timestamp: tsMark,
					Value:     float64(pair.Value / 1000),
				})
				tsMark += tRange.Step.Microseconds()
			}

			ts.LatencyP90 = append(ts.LatencyP90, MetricsPoint{
				Timestamp: int64(pair.Timestamp) * 1000,
				Value:     float64(pair.Value / 1000),
			})
			tsMark += tRange.Step.Microseconds()
		}

		if tsMark < tRange.End.UnixMicro() {
			for tsMark < tRange.End.UnixMicro() {
				ts.LatencyP90 = append(ts.LatencyP90, MetricsPoint{
					Timestamp: tsMark,
					Value:     float64(0),
				})
				tsMark += tRange.Step.Microseconds()
			}
		}

		result = append(result, ts)
	}

	return result
}
