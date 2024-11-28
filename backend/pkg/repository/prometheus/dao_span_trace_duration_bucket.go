package prometheus

import (
	"context"
	"fmt"
	"time"

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

// 基于服务列表、URL列表和时段、步长，查询P90曲线
func (repo *promRepo) QueryRangePercentile(startTime int64, endTime int64, step int64, nodes *model.TopologyNodes) ([]DescendantMetrics, error) {
	svcs, endpoints, _ := nodes.GetLabels(model.GROUP_SERVICE)
	if len(svcs) == 0 {
		return nil, nil
	}

	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}
	query := getSpanTraceP9xSql(repo.GetRange(), tRange.Step, svcs, endpoints)
	res, _, err := repo.GetApi().QueryRange(context.Background(), query, tRange)
	if err != nil {
		return nil, err
	}

	return getDescendantMetrics("svc_name", "content_key", tRange, res), nil
}

func getSpanTraceP9xSql(promRange string, step time.Duration, svcs []string, endpoints []string) string {
	builder := NewUnionP9xBuilder(
		"0.9",
		"kindling_span_trace_duration_nanoseconds_bucket",
		[]string{promRange, "content_key", "svc_name"},
		step,
	)
	builder.AddCondition("svc_name", svcs)
	builder.AddCondition("content_key", endpoints)
	return builder.ToString()
}

func (repo *promRepo) QueryInstanceP90(startTime int64, endTime int64, step int64, endpoint string, instance *model.ServiceInstance) (map[int64]float64, error) {
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
		// VM场景
		extraCondition = fmt.Sprintf("node_name='%s', pid='%d'", instance.NodeName, instance.Pid)
	}
	sql := getSpanTraceInstanceP9xSql(repo.GetRange(), tRange.Step, endpoint, extraCondition)

	res, _, err := repo.GetApi().QueryRange(context.Background(), sql, tRange)
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
			result[int64(pair.Timestamp)*1000] = float64(pair.Value / 1000) // 返回时间为us
		}
	}
	return result, nil
}

func getSpanTraceInstanceP9xSql(promRange string, step time.Duration, endpoint string, extraCondition string) string {
	builder := NewUnionP9xBuilder(
		"0.9",
		"kindling_span_trace_duration_nanoseconds_bucket",
		[]string{promRange},
		step,
	)
	builder.AddCondition("content_key", []string{endpoint})
	builder.AddExtraCondition(extraCondition)

	return builder.ToString()
}

type DescendantMetrics struct {
	ServiceName string         `json:"serviceName"` // 服务名
	EndPoint    string         `json:"endpoint"`    // Endpoint
	LatencyP90  []MetricsPoint `json:"latencyP90"`  // P90曲线值
}

type MetricsPoint struct {
	Timestamp int64   `json:"timestamp"` // 时间(微秒)
	Value     float64 `json:"value"`     // 值
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
			// 未获取到数据的时间点填充0
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
