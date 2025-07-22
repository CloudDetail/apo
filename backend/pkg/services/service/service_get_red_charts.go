// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"log"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
)

func (s *service) GetServiceREDCharts(ctx core.Context, req *request.GetServiceREDChartsRequest) (response.GetServiceREDChartsResponse, error) {
	step := time.Duration(req.Step * 1000)
	// filters := make([]string, 0, 4)

	var err error

	var filter prometheus.PQLFilter
	if req.GroupID > 0 {
		filter, err = common.GetPQLFilterByGroupID(ctx, s.dbRepo, "apm", req.GroupID)
		if err != nil {
			return response.GetServiceREDChartsResponse{}, err
		}
	} else {
		filter = prometheus.NewFilter()
	}

	if len(req.ClusterIDs) > 0 {
		filter.RegexMatch(prometheus.ClusterIDKey, prometheus.RegexMultipleValue(req.ClusterIDs...))
	}
	if len(req.ServiceList) > 0 {
		filter.RegexMatch(prometheus.ServiceNameKey, prometheus.RegexMultipleValue(req.ServiceList...))
	}
	if len(req.EndpointList) > 0 {
		filter.RegexMatch(prometheus.ContentKeyKey, prometheus.RegexMultipleValue(req.EndpointList...))
	}

	serviceDetails := make(map[string]map[string]response.RedCharts)
	latencyRes, err := s.promRepo.QueryRangeMetricsWithPQLFilter(ctx,
		prometheus.PQLAvgLatencyWithPQLFilter,
		req.StartTime,
		req.EndTime,
		step.Microseconds(),
		prometheus.EndpointGranularity,
		filter,
	)
	if err != nil {
		log.Println("get instance range data error: ", err)
	}

	mergeResult(latencyRes, serviceDetails, "latency")

	errorRes, err := s.promRepo.QueryRangeMetricsWithPQLFilter(ctx,
		prometheus.PQLAvgErrorRateWithPQLFilter,
		req.StartTime,
		req.EndTime,
		step.Microseconds(),
		prometheus.EndpointGranularity,
		filter,
	)

	mergeResult(errorRes, serviceDetails, "errorRate")

	if err != nil {
		log.Println("get instance range data error: ", err)
	}

	tpsRes, err := s.promRepo.QueryRangeMetricsWithPQLFilter(ctx,
		prometheus.PQLAvgTPSWithPQLFilter,
		req.StartTime,
		req.EndTime,
		step.Microseconds(),
		prometheus.EndpointGranularity,
		filter,
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
