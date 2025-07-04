// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"math"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

const (
	metricLatencyData = "latencyData"
	metricErrorData   = "errorData"
	metricTPMData     = "TPMData"
)

type ServiceMap = prometheus.MetricGroupMap[prometheus.ServiceKey, *prometheus.ServiceMetrics]

func (s *service) GetServiceRedCharts(ctx core.Context, req *request.QueryServiceRedChartsRequest) *response.QueryServiceRedChartsResponse {
	serviceKey := prometheus.ServiceKey{
		SvcName: req.ServiceName,
	}
	serviceMap := &ServiceMap{
		MetricGroupList: []*prometheus.ServiceMetrics{},
		MetricGroupMap: map[prometheus.ServiceKey]*prometheus.ServiceMetrics{
			serviceKey: {
				ServiceKey: serviceKey,
			},
		},
	}
	startTime := time.Unix(req.StartTime/1000000, 0)
	endTime := time.Unix(req.EndTime/1000000, 0)
	granularity := prometheus.SVCGranularity
	filters := []string{prometheus.ServicePQLFilter, req.ServiceName}
	// Chart data
	stepMs := getStepMs(req.EndTime - req.StartTime)
	latencyRes, latencyErr := s.promRepo.QueryRangeAggMetricsWithFilter(ctx,
		prometheus.PQLAvgLatencyWithFilters,
		req.StartTime,
		req.EndTime,
		stepMs,
		granularity,
		filters...,
	)
	if latencyErr != nil {
		return &response.QueryServiceRedChartsResponse{
			Msg: "query avg metrics failed: " + latencyErr.Error(),
		}
	}

	if len(latencyRes) == 0 {
		return s.queryServiceRedsByApi(ctx, req)
	}
	mergeServiceChartMetrics(serviceMap, latencyRes, metricLatencyData)

	errorRes, rateErr := s.promRepo.QueryRangeAggMetricsWithFilter(ctx,
		prometheus.PQLAvgErrorRateWithFilters,
		req.StartTime,
		req.EndTime,
		stepMs,
		granularity,
		filters...,
	)
	if rateErr != nil {
		return &response.QueryServiceRedChartsResponse{
			Msg: "query errorRate metrics failed: " + rateErr.Error(),
		}
	}
	mergeServiceChartMetrics(serviceMap, errorRes, metricErrorData)

	tpmRes, tmpErr := s.promRepo.QueryRangeAggMetricsWithFilter(ctx,
		prometheus.PQLAvgTPSWithFilters,
		req.StartTime,
		req.EndTime,
		stepMs,
		granularity,
		filters...,
	)
	if tmpErr != nil {
		return &response.QueryServiceRedChartsResponse{
			Msg: "query tps metrics failed: " + tmpErr.Error(),
		}
	}
	mergeServiceChartMetrics(serviceMap, tpmRes, metricTPMData)

	// Metric Value
	s.promRepo.FillMetric(ctx, serviceMap, prometheus.AVG, startTime, endTime, filters, granularity)
	s.promRepo.FillMetric(ctx, serviceMap, prometheus.DOD, startTime, endTime, filters, granularity)
	s.promRepo.FillMetric(ctx, serviceMap, prometheus.WOW, startTime, endTime, filters, granularity)

	results := make([]*response.QueryChartResult, 0)
	for _, serviceMetric := range serviceMap.MetricGroupMap {
		latencyTempChartObject := response.TempChartObject{
			Ratio: response.Ratio{
				DayOverDay:  serviceMetric.REDMetrics.DOD.Latency,
				WeekOverDay: serviceMetric.REDMetrics.WOW.Latency,
			},
			Value: serviceMetric.REDMetrics.Avg.Latency,
		}
		if serviceMetric.LatencyData != nil {
			latencyTempChartObject.ChartData = DataToChart(serviceMetric.LatencyData)
		} else {
			latencyTempChartObject.ChartData = FillChart(startTime, endTime, stepMs)
		}
		results = append(results, &response.QueryChartResult{
			Title: "平均响应时间",
			Unit:  "ms",
			Timeseries: []*response.Timeseries{
				{
					Legend:       req.ServiceName,
					LegendFormat: "",
					Labels: map[string]string{
						"service": req.ServiceName,
					},
					Chart: latencyTempChartObject,
				},
			},
		})

		errorTempChartObject := response.TempChartObject{
			Ratio: response.Ratio{
				DayOverDay:  serviceMetric.REDMetrics.DOD.ErrorRate,
				WeekOverDay: serviceMetric.REDMetrics.WOW.ErrorRate,
			},
			Value: serviceMetric.REDMetrics.Avg.ErrorRate,
		}
		if errorTempChartObject.Value == nil {
			zero := new(float64)
			errorTempChartObject.Value = zero
		}
		if serviceMetric.ErrorRateData != nil {
			errorTempChartObject.ChartData = DataToChart(serviceMetric.ErrorRateData)
		} else {
			errorTempChartObject.ChartData = FillChart(startTime, endTime, stepMs)
		}
		results = append(results, &response.QueryChartResult{
			Title: "错误率",
			Unit:  "%",
			Timeseries: []*response.Timeseries{
				{
					Legend:       req.ServiceName,
					LegendFormat: "",
					Labels: map[string]string{
						"service": req.ServiceName,
					},
					Chart: errorTempChartObject,
				},
			},
		})

		// construct tps return value
		tpmTempChartObject := response.TempChartObject{
			Ratio: response.Ratio{
				DayOverDay:  serviceMetric.REDMetrics.DOD.TPM,
				WeekOverDay: serviceMetric.REDMetrics.WOW.TPM,
			},
			Value: serviceMetric.REDMetrics.Avg.TPM,
		}
		if serviceMetric.TPMData != nil {
			tpmTempChartObject.ChartData = DataToChart(serviceMetric.TPMData)
		} else {
			tpmTempChartObject.ChartData = FillChart(startTime, endTime, stepMs)
		}
		results = append(results, &response.QueryChartResult{
			Title: "吞吐量",
			Unit:  "次/分",
			Timeseries: []*response.Timeseries{
				{
					Legend:       req.ServiceName,
					LegendFormat: "",
					Labels: map[string]string{
						"service": req.ServiceName,
					},
					Chart: tpmTempChartObject,
				},
			},
		})
	}

	return &response.QueryServiceRedChartsResponse{
		Results: results,
	}
}

func mergeServiceChartMetrics(serviceMap *ServiceMap, results []prometheus.MetricResult, metricName string) {
	for _, res := range results {
		var kType prometheus.ServiceKey
		key := kType.ConvertFromLabels(res.Metric).(prometheus.ServiceKey)

		service, ok := serviceMap.MetricGroupMap[key]
		if !ok {
			continue
		}
		switch metricName {
		case metricLatencyData:
			for i := range res.Values {
				res.Values[i].Value /= 1e3
			}
			service.LatencyData = res.Values
		case metricErrorData:
			for i := range res.Values {
				res.Values[i].Value *= 100
			}
			service.ErrorRateData = res.Values
		case metricTPMData:
			for i := range res.Values {
				res.Values[i].Value *= 60
			}
			service.TPMData = res.Values
		}
	}
}

func getStepMs(duration int64) int64 {
	var stepMs int64
	if duration <= 3600_000_000 {
		stepMs = 60_000_000 // 1m
	} else if duration <= 18_000_000_000 {
		stepMs = 300_000_000 // 5m
	} else if duration <= 36_000_000_000 {
		stepMs = 600_000_000 // 10m
	} else if duration <= 108_000_000_000 {
		stepMs = 18_000_000_000 // 30m
	} else if duration <= 216_000_000_000 {
		stepMs = 36_000_000_000 // 2h
	} else {
		stepMs = 86_400_000_000 // 1d
	}
	return stepMs
}

// DataToChart convert chart data to map
func DataToChart(data []prometheus.Points) map[int64]float64 {
	chart := make(map[int64]float64)
	for _, item := range data {
		timestamp := item.TimeStamp
		value := item.Value
		if !math.IsInf(value, 1) {
			chart[timestamp] = value
		} else {
			chart[timestamp] = prometheus.RES_MAX_VALUE
		}
	}
	return chart
}

func FillChart(startTime, endTime time.Time, stepMs int64) map[int64]float64 {
	values := make(map[int64]float64)
	for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += stepMs {
		values[ts] = 0
	}
	return values
}

func (s *service) queryServiceRedsByApi(ctx core.Context, req *request.QueryServiceRedChartsRequest) *response.QueryServiceRedChartsResponse {
	return &response.QueryServiceRedChartsResponse{
		Msg: "Data not found",
	}
}
