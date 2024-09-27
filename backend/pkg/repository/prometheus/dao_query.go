package prometheus

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func (repo *promRepo) QueryData(searchTime time.Time, query string) ([]MetricResult, error) {
	value, warnings, err := repo.GetApi().Query(context.Background(), query, searchTime)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Println("Warnings:", warnings)
	}
	var results []MetricResult
	vector, ok := value.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Vector", value)
	}

	for _, sample := range vector {
		metric := Labels{}
		for name, value := range sample.Metric {
			switch string(name) {
			case "container_id":
				metric.ContainerID = string(value)
			case "content_key":
				metric.ContentKey = string(value)
			case "instance":
				metric.Instance = string(value)
			case "is_error":
				metric.IsError = string(value)
			case "job":
				metric.Job = string(value)
			case "node_name":
				metric.NodeName = string(value)
			case "pod":
				metric.POD = string(value)
			case "svc_name":
				metric.SvcName = string(value)
			case "top_span":
				metric.TopSpan = string(value)
			case "pid":
				metric.PID = string(value)
			case "pod_name":
				metric.PodName = string(value)
			case "namespace":
				metric.Namespace = string(value)
			case "db_system":
				metric.DBSystem = string(value)
			case "db_name":
				metric.DBName = string(value)
			case "name":
				metric.Name = string(value)
			case "db_url":
				metric.DBUrl = string(value)
			}
		}

		values := []Points{
			{sample.Timestamp.UnixNano() / 1e3, float64(sample.Value)},
		}

		results = append(results, MetricResult{
			Metric: metric,
			Values: values,
		})
	}
	return results, nil
}
func (repo *promRepo) QueryRangeData(startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error) {
	r := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	}

	value, warnings, err := repo.GetApi().QueryRange(context.Background(), query, r)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Println("Warnings:", warnings)
	}

	var results []MetricResult
	matrix, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Matrix", value)
	}

	// Process the result matrix
	for _, stream := range matrix {
		metric := Labels{}
		for name, value := range stream.Metric {
			switch string(name) {
			case "container_id":
				metric.ContainerID = string(value)
			case "content_key":
				metric.ContentKey = string(value)
			case "instance":
				metric.Instance = string(value)
			case "is_error":
				metric.IsError = string(value)
			case "job":
				metric.Job = string(value)
			case "node_name":
				metric.NodeName = string(value)
			case "pod":
				metric.POD = string(value)
			case "svc_name":
				metric.SvcName = string(value)
			case "top_span":
				metric.TopSpan = string(value)
			case "pid":
				metric.PID = string(value)
			case "pod_name":
				metric.PodName = string(value)
			case "namespace":
				metric.Namespace = string(value)
			case "db_system":
				metric.DBSystem = string(value)
			case "db_name":
				metric.DBName = string(value)
			case "name":
				metric.Name = string(value)
			case "db_url":
				metric.DBUrl = string(value)
			}
		}

		var values []Points
		for _, point := range stream.Values {
			values = append(values, Points{point.Timestamp.UnixNano() / 1e3, float64(point.Value)})
		}

		results = append(results, MetricResult{
			Metric: metric,
			Values: values,
		})
	}

	return results, nil
}

//latency查询曲线图需要处理，将其转成微秒

func (repo *promRepo) QueryLatencyData(searchTime time.Time, query string) ([]MetricResult, error) {
	value, warnings, err := repo.GetApi().Query(context.Background(), query, searchTime)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Println("Warnings:", warnings)
	}
	var results []MetricResult
	vector, ok := value.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Vector", value)
	}

	for _, sample := range vector {
		metric := Labels{}
		for name, value := range sample.Metric {
			switch string(name) {
			case "container_id":
				metric.ContainerID = string(value)
			case "content_key":
				metric.ContentKey = string(value)
			case "instance":
				metric.Instance = string(value)
			case "is_error":
				metric.IsError = string(value)
			case "job":
				metric.Job = string(value)
			case "node_name":
				metric.NodeName = string(value)
			case "pod":
				metric.POD = string(value)
			case "svc_name":
				metric.SvcName = string(value)
			case "top_span":
				metric.TopSpan = string(value)
			case "pid":
				metric.PID = string(value)
			case "pod_name":
				metric.PodName = string(value)
			case "namespace":
				metric.Namespace = string(value)
			case "node_ip":
				metric.NodeIP = string(value)
			}
		}

		values := []Points{
			{sample.Timestamp.UnixNano() / 1e3, float64(sample.Value) / 1e3},
		}

		results = append(results, MetricResult{
			Metric: metric,
			Values: values,
		})
	}
	return results, nil
}

func (repo *promRepo) QueryRangeLatencyData(startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error) {
	r := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	}

	value, warnings, err := repo.GetApi().QueryRange(context.Background(), query, r)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Println("Warnings:", warnings)
	}

	var results []MetricResult
	matrix, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Matrix", value)
	}

	// Process the result matrix
	for _, stream := range matrix {
		metric := Labels{}
		for name, value := range stream.Metric {
			switch string(name) {
			case "container_id":
				metric.ContainerID = string(value)
			case "content_key":
				metric.ContentKey = string(value)
			case "instance":
				metric.Instance = string(value)
			case "is_error":
				metric.IsError = string(value)
			case "job":
				metric.Job = string(value)
			case "node_name":
				metric.NodeName = string(value)
			case "pod":
				metric.POD = string(value)
			case "svc_name":
				metric.SvcName = string(value)
			case "top_span":
				metric.TopSpan = string(value)
			case "pid":
				metric.PID = string(value)
			case "pod_name":
				metric.PodName = string(value)
			case "namespace":
				metric.Namespace = string(value)
			}
		}

		var values []Points
		for _, point := range stream.Values {
			values = append(values, Points{point.Timestamp.UnixNano() / 1e3, float64(point.Value / 1e3)})
		}

		results = append(results, MetricResult{
			Metric: metric,
			Values: values,
		})
	}

	return results, nil
}

// 错误率需要处理为百分比
func (repo *promRepo) QueryErrorRateData(searchTime time.Time, query string) ([]MetricResult, error) {
	value, warnings, err := repo.GetApi().Query(context.Background(), query, searchTime)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Println("Warnings:", warnings)
	}
	var results []MetricResult
	vector, ok := value.(model.Vector)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Vector", value)
	}

	for _, sample := range vector {
		metric := Labels{}
		for name, value := range sample.Metric {
			switch string(name) {
			case "container_id":
				metric.ContainerID = string(value)
			case "content_key":
				metric.ContentKey = string(value)
			case "instance":
				metric.Instance = string(value)
			case "is_error":
				metric.IsError = string(value)
			case "job":
				metric.Job = string(value)
			case "node_name":
				metric.NodeName = string(value)
			case "pod":
				metric.POD = string(value)
			case "svc_name":
				metric.SvcName = string(value)
			case "top_span":
				metric.TopSpan = string(value)
			case "pid":
				metric.PID = string(value)
			case "pod_name":
				metric.PodName = string(value)
			case "namespace":
				metric.Namespace = string(value)
			}
		}

		values := []Points{
			{sample.Timestamp.UnixNano() / 1e3, float64(sample.Value) * 100},
		}

		results = append(results, MetricResult{
			Metric: metric,
			Values: values,
		})
	}
	return results, nil
}

func (repo *promRepo) QueryRangeErrorData(startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error) {
	r := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	}

	value, warnings, err := repo.GetApi().QueryRange(context.Background(), query, r)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Println("Warnings:", warnings)
	}

	var results []MetricResult
	matrix, ok := value.(model.Matrix)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T, expected model.Matrix", value)
	}

	// Process the result matrix
	for _, stream := range matrix {
		metric := Labels{}
		for name, value := range stream.Metric {
			switch string(name) {
			case "container_id":
				metric.ContainerID = string(value)
			case "content_key":
				metric.ContentKey = string(value)
			case "instance":
				metric.Instance = string(value)
			case "is_error":
				metric.IsError = string(value)
			case "job":
				metric.Job = string(value)
			case "node_name":
				metric.NodeName = string(value)
			case "pod":
				metric.POD = string(value)
			case "svc_name":
				metric.SvcName = string(value)
			case "top_span":
				metric.TopSpan = string(value)
			case "pid":
				metric.PID = string(value)
			case "pod_name":
				metric.PodName = string(value)
			case "namespace":
				metric.Namespace = string(value)
			}
		}

		var values []Points
		for _, point := range stream.Values {
			values = append(values, Points{point.Timestamp.UnixNano() / 1e3, float64(point.Value * 100)})
		}

		results = append(results, MetricResult{
			Metric: metric,
			Values: values,
		})
	}

	return results, nil
}
