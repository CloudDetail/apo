// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

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
