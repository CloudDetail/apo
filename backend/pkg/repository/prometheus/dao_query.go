// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"log"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func (repo *promRepo) QueryData(ctx core.Context, searchTime time.Time, query string) ([]MetricResult, error) {
	value, warnings, err := repo.GetApi().Query(ctx.GetContext(), query, searchTime)
	if err != nil {
		return nil, fmt.Errorf("query metric failed, err: %w, query: %s, timestamp: %d", err, query, searchTime.Unix())
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
		metric.Extract(sample.Metric)
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

func (repo *promRepo) QueryRangeData(ctx core.Context, startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error) {
	r := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	}

	value, warnings, err := repo.GetApi().QueryRange(ctx.GetContext(), query, r)
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
		metric.Extract(stream.Metric)

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

// latency the query graph needs to be processed, turn it into microseconds

func (repo *promRepo) QueryLatencyData(ctx core.Context, searchTime time.Time, query string) ([]MetricResult, error) {
	value, warnings, err := repo.GetApi().Query(ctx.GetContext(), query, searchTime)
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
		metric.Extract(sample.Metric)

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

func (repo *promRepo) QueryRangeLatencyData(ctx core.Context, startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error) {
	r := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	}

	value, warnings, err := repo.GetApi().QueryRange(ctx.GetContext(), query, r)
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
		metric.Extract(stream.Metric)

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

// Error rate needs to be handled as a percentage
func (repo *promRepo) QueryErrorRateData(ctx core.Context, searchTime time.Time, query string) ([]MetricResult, error) {
	value, warnings, err := repo.GetApi().Query(ctx.GetContext(), query, searchTime)
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
		metric.Extract(sample.Metric)

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

func (repo *promRepo) QueryRangeErrorData(ctx core.Context, startTime time.Time, endTime time.Time, query string, step time.Duration) ([]MetricResult, error) {
	r := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  step,
	}

	value, warnings, err := repo.GetApi().QueryRange(ctx.GetContext(), query, r)
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
		metric.Extract(stream.Metric)

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
