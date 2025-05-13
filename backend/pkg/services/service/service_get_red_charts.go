// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"log"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

func (s *service) GetServiceREDCharts(ctx_core core.Context, req *request.GetServiceREDChartsRequest) (response.GetServiceREDChartsResponse, error) {
	step := time.Duration(req.Step * 1000)
	filters := make([]string, 0, 4)
	filters = append(filters, prometheus.ServiceRegexPQLFilter)
	filters = append(filters, prometheus.RegexMultipleValue(req.ServiceList...))
	filters = append(filters, prometheus.ContentKeyRegexPQLFilter)
	filters = append(filters, prometheus.RegexMultipleValue(req.EndpointList...))
	serviceDetails := make(map[string]map[string]response.RedCharts)
	latencyRes, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgLatencyWithFilters,
		req.StartTime,
		req.EndTime,
		step.Microseconds(),
		prometheus.EndpointGranularity,
		filters...,
	)
	if err != nil {
		log.Println("get instance range data error: ", err)
	}

	mergeResult(latencyRes, serviceDetails, "latency")

	errorRes, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgErrorRateWithFilters,
		req.StartTime,
		req.EndTime,
		step.Microseconds(),
		prometheus.EndpointGranularity,
		filters...,
	)

	mergeResult(errorRes, serviceDetails, "errorRate")

	if err != nil {
		log.Println("get instance range data error: ", err)
	}

	tpsRes, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgTPSWithFilters,
		req.StartTime,
		req.EndTime,
		step.Microseconds(),
		prometheus.EndpointGranularity,
		filters...,
	)

	mergeResult(tpsRes, serviceDetails, "rps")

	if err != nil {
		log.Println("get instance range data error: ", err)
	}
	return serviceDetails, nil
}

func mergeResult(results []prometheus.MetricResult, serviceDetails map[string]map[string]response.RedCharts, metricType string) {
	for _, result := range results {
		svc, contentKey := result.Metric.SvcName, result.Metric.ContentKey
		contents, ok := serviceDetails[svc]
		if !ok {
			contents = make(map[string]response.RedCharts)
			serviceDetails[svc] = contents
		}

		redCharts := contents[contentKey]

		switch metricType {
		case "latency":
			for i := range result.Values {
				result.Values[i].Value /= 1e3
			}
			redCharts.Latency = DataToChart(result.Values)
		case "errorRate":
			for i := range result.Values {
				result.Values[i].Value *= 100
			}
			redCharts.ErrorRate = DataToChart(result.Values)
		case "rps":
			for i := range result.Values {
				result.Values[i].Value *= 60
			}
			redCharts.RPS = DataToChart(result.Values)
		}

		contents[contentKey] = redCharts
	}
}
