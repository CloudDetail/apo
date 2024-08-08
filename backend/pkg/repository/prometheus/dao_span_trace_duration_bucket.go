package prometheus

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prometheus_model "github.com/prometheus/common/model"
)

const (
	TEMPLATE_FILTER_ENDPOINT            = `content_key=~"%s"`
	TEMPLATE_FILTER_SVC                 = `svc_name=~"%s"`
	TEMPLATE_HISTO_P90_LATENCY          = `histogram_quantile(0.9,sum by (%s,content_key,svc_name) (increase(kindling_span_trace_duration_nanoseconds_bucket{%s}[%s])))`
	TEMPLATE_INSTANCE_HISTO_P90_LATENCY = `histogram_quantile(0.9,sum by (%s) (increase(kindling_span_trace_duration_nanoseconds_bucket{%s, content_key='%s'}[%s])))`

	TIME_SECOND = int64(time.Second)
	TIME_MINUTE = 60 * TIME_SECOND
	TIME_HOUR   = 60 * TIME_MINUTE
	TIME_DAY    = TIME_HOUR * 24
)

// 基于服务列表、URL列表和时段、步长，查询P90曲线
func (repo *promRepo) QueryRangePercentile(startTime int64, endTime int64, step int64, services []string, endpoints []string) ([]response.GetDescendantMetricsResponse, error) {
	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}

	filters := []string{}
	if len(endpoints) > 0 {
		filters = append(filters, fmt.Sprintf(TEMPLATE_FILTER_ENDPOINT, strings.Join(endpoints, "|")))
	}
	if len(services) > 0 {
		filters = append(filters, fmt.Sprintf(TEMPLATE_FILTER_SVC, strings.Join(services, "|")))
	}

	query := fmt.Sprintf(TEMPLATE_HISTO_P90_LATENCY,
		repo.GetRange(),
		strings.Join(filters, ","),
		getDurationFromStep(tRange.Step),
	)
	res, _, err := repo.GetApi().QueryRange(context.Background(), query, tRange)
	if err != nil {
		return nil, err
	}

	result := make([]response.GetDescendantMetricsResponse, 0)
	values, ok := res.(prometheus_model.Matrix)
	if !ok {
		return result, nil
	}

	for _, val := range values {
		svcName, ok := val.Metric["svc_name"]
		if !ok {
			continue
		}
		contentKey, ok := val.Metric["content_key"]
		if !ok {
			continue
		}

		ts := response.GetDescendantMetricsResponse{
			ServiceName: string(svcName),
			EndPoint:    string(contentKey),
			LatencyP90:  []response.MetricsPoint{},
		}

		tsMark := tRange.Start.UnixMicro()
		for _, pair := range val.Values {
			// 未获取到数据的时间点填充0
			for tsMark < int64(pair.Timestamp)*1000 {
				ts.LatencyP90 = append(ts.LatencyP90, response.MetricsPoint{
					Timestamp: tsMark,
					Value:     float64(pair.Value / 1000),
				})
				tsMark += tRange.Step.Microseconds()
			}

			ts.LatencyP90 = append(ts.LatencyP90, response.MetricsPoint{
				Timestamp: int64(pair.Timestamp) * 1000,
				Value:     float64(pair.Value / 1000),
			})
			tsMark += tRange.Step.Microseconds()
		}

		if tsMark < tRange.End.UnixMicro() {
			for tsMark < tRange.End.UnixMicro() {
				ts.LatencyP90 = append(ts.LatencyP90, response.MetricsPoint{
					Timestamp: tsMark,
					Value:     float64(0),
				})
				tsMark += tRange.Step.Microseconds()
			}
		}

		result = append(result, ts)
	}

	return result, nil
}

func (repo *promRepo) QueryInstanceP90(startTime int64, endTime int64, step int64, endpoint string, instance *model.ServiceInstance) (map[int64]float64, error) {
	tRange := v1.Range{
		Start: time.UnixMicro(startTime),
		End:   time.UnixMicro(endTime),
		Step:  time.Duration(step * 1000),
	}

	var queryCondition string
	if instance.PodName != "" {
		queryCondition = fmt.Sprintf("pod='%s'", instance.PodName)
	} else if instance.ContainerId != "" {
		queryCondition = fmt.Sprintf("node_name='%s', container_id='%s'", instance.NodeName, instance.ContainerId)
	} else {
		// VM场景
		queryCondition = fmt.Sprintf("node_name='%s', pid='%d'", instance.NodeName, instance.Pid)
	}

	query := fmt.Sprintf(TEMPLATE_INSTANCE_HISTO_P90_LATENCY,
		repo.GetRange(),
		queryCondition,
		endpoint,
		getDurationFromStep(tRange.Step),
	)
	res, _, err := repo.GetApi().QueryRange(context.Background(), query, tRange)
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

func getDurationFromStep(step time.Duration) string {
	var stepNS = step.Nanoseconds()
	if stepNS > TIME_DAY && (stepNS%TIME_DAY == 0) {
		return strconv.FormatInt(stepNS/TIME_DAY, 10) + "d"
	}
	if stepNS > TIME_HOUR && (stepNS%TIME_HOUR == 0) {
		return strconv.FormatInt(stepNS/TIME_HOUR, 10) + "h"
	}
	if stepNS > TIME_MINUTE && (stepNS%TIME_MINUTE == 0) {
		return strconv.FormatInt(stepNS/TIME_MINUTE, 10) + "m"
	}
	if stepNS > TIME_SECOND && (stepNS%TIME_SECOND == 0) {
		return strconv.FormatInt(stepNS/TIME_SECOND, 10) + "s"
	}

	// 默认时间
	return "1m"
}
