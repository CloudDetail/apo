// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"math"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/util"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

const (
	metricLatencyData = "latencyData"
	metricErrorData   = "errorData"
	metricTPMData     = "TPMData"
)

type ServiceMap = prometheus.MetricGroupMap[prometheus.ServiceKey, *prometheus.ServiceMetrics]

func (s *service) GetServiceRedCharts(ctx core.Context, req *request.QueryServiceRedChartsRequest) *response.QueryServiceRedChartsResponse {
	stepMicros := getStepMicros(req.EndTime - req.StartTime)
	startBucket := req.StartTime / stepMicros
	endBucket := req.EndTime / stepMicros
	results := make([]*response.QueryChartResult, 0)
	bucketMetrics, err := s.chRepo.QueryGroupServiceRedMetrics(ctx, req.StartTime, req.EndTime, req.Cluster, req.ServiceName, "", stepMicros)
	if err != nil {
		return &response.QueryServiceRedChartsResponse{
			Msg: "query redmetrics: " + err.Error(),
		}
	}
	if len(bucketMetrics) == 0 {
		return &response.QueryServiceRedChartsResponse{
			Results: results,
		}
	}
	bucketMap := make(map[int64]*clickhouse.BucketRedMetric)
	for _, bucketMetric := range bucketMetrics {
		bucketMap[int64(bucketMetric.TimeBucket)] = &bucketMetric
	}

	countChart := make(map[int64]float64)
	errorChart := make(map[int64]float64)
	latencyChart := make(map[int64]float64)
	groupMetrics := &clickhouse.GroupRedMetric{
		Duration: req.EndTime - req.StartTime,
	}
	for i := startBucket; i <= endBucket; i++ {
		bucketMetric, ok := bucketMap[i]
		if !ok {
			// Fill Zero
			countChart[i*stepMicros] = 0.0
			errorChart[i*stepMicros] = 0.0
			latencyChart[i*stepMicros] = 0.0
		} else {
			if bucketMetric.TotalCount == 0 {
				countChart[i*stepMicros] = 0.0
				errorChart[i*stepMicros] = 0.0
				latencyChart[i*stepMicros] = 0.0
			} else {
				countChart[i*stepMicros] = float64(int64(bucketMetric.TotalCount) * 60_000_000 / stepMicros)
				errorChart[i*stepMicros] = float64(100 * bucketMetric.TotalError / bucketMetric.TotalCount)
				latencyChart[i*stepMicros] = float64(bucketMetric.TotalDuration / uint64(bucketMetric.TotalCount))

				groupMetrics.TotalCount += bucketMetric.TotalCount
				groupMetrics.TotalError += bucketMetric.TotalError
				groupMetrics.TotalDuration += bucketMetric.TotalDuration
			}
		}
	}

	dodMetrics, err := s.chRepo.QueryGroupServiceRedMetricValue(ctx, req.StartTime-24*int64(time.Hour)/1000, req.StartTime, req.Cluster, req.ServiceName, req.Endpoint)
	if err != nil {
		return &response.QueryServiceRedChartsResponse{
			Msg: "query dod metric failed: " + err.Error(),
		}
	}
	wowMetrics, err := s.chRepo.QueryGroupServiceRedMetricValue(ctx, req.StartTime-7*24*int64(time.Hour)/1000, req.StartTime, req.Cluster, req.ServiceName, req.Endpoint)
	if err != nil {
		return &response.QueryServiceRedChartsResponse{
			Msg: "query wow metric failed: " + err.Error(),
		}
	}
	avgLatency := groupMetrics.GetAvgLatency()
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
				Chart: response.TempChartObject{
					ChartData: latencyChart,
					Ratio: response.Ratio{
						DayOverDay:  util.PtrFloat64(GetRatio(avgLatency, dodMetrics.GetAvgLatency())),
						WeekOverDay: util.PtrFloat64(GetRatio(avgLatency, wowMetrics.GetAvgLatency())),
					},
					Value: util.PtrFloat64(avgLatency),
				},
			},
		},
	})
	errorRate := groupMetrics.GetErrorRate()
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
				Chart: response.TempChartObject{
					ChartData: errorChart,
					Ratio: response.Ratio{
						DayOverDay:  util.PtrFloat64(GetRatio(errorRate, dodMetrics.GetErrorRate())),
						WeekOverDay: util.PtrFloat64(GetRatio(errorRate, wowMetrics.GetErrorRate())),
					},
					Value: util.PtrFloat64(errorRate),
				},
			},
		},
	})
	tpm := groupMetrics.GetTpm()
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
				Chart: response.TempChartObject{
					ChartData: countChart,
					Ratio: response.Ratio{
						DayOverDay:  util.PtrFloat64(GetRatio(tpm, dodMetrics.GetTpm())),
						WeekOverDay: util.PtrFloat64(GetRatio(tpm, wowMetrics.GetTpm())),
					},
					Value: util.PtrFloat64(tpm),
				},
			},
		},
	})
	return &response.QueryServiceRedChartsResponse{
		Results: results,
	}
}

func (s *service) getServiceRedChartsByApo(ctx core.Context, req *request.QueryServiceRedChartsRequest) *response.QueryServiceRedChartsResponse {
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

	filter := prometheus.NewFilter()
	filter.Equal(prometheus.ServiceNameKey, req.ServiceName)
	// Chart data
	stepMicros := getStepMicros(req.EndTime - req.StartTime)
	latencyRes, latencyErr := s.promRepo.QueryRangeMetricsWithPQLFilter(ctx,
		prometheus.PQLAvgLatencyWithPQLFilter,
		req.StartTime,
		req.EndTime,
		stepMicros,
		granularity,
		filter,
	)
	if latencyErr != nil {
		return &response.QueryServiceRedChartsResponse{
			Msg: "query avg metrics failed: " + latencyErr.Error(),
		}
	}

	if len(latencyRes) == 0 {
		return &response.QueryServiceRedChartsResponse{
			Results: []*response.QueryChartResult{},
		}
	}
	mergeServiceChartMetrics(serviceMap, latencyRes, metricLatencyData)

	errorRes, rateErr := s.promRepo.QueryRangeMetricsWithPQLFilter(ctx,
		prometheus.PQLAvgErrorRateWithPQLFilter,
		req.StartTime,
		req.EndTime,
		stepMicros,
		granularity,
		filter,
	)
	if rateErr != nil {
		return &response.QueryServiceRedChartsResponse{
			Msg: "query errorRate metrics failed: " + rateErr.Error(),
		}
	}
	mergeServiceChartMetrics(serviceMap, errorRes, metricErrorData)

	tpmRes, tmpErr := s.promRepo.QueryRangeMetricsWithPQLFilter(ctx,
		prometheus.PQLAvgTPSWithPQLFilter,
		req.StartTime,
		req.EndTime,
		stepMicros,
		granularity,
		filter,
	)
	if tmpErr != nil {
		return &response.QueryServiceRedChartsResponse{
			Msg: "query tps metrics failed: " + tmpErr.Error(),
		}
	}
	mergeServiceChartMetrics(serviceMap, tpmRes, metricTPMData)

	// Metric Value
	s.promRepo.FillMetric(ctx, serviceMap, prometheus.AVG, startTime, endTime, filter, granularity)
	s.promRepo.FillMetric(ctx, serviceMap, prometheus.DOD, startTime, endTime, filter, granularity)
	s.promRepo.FillMetric(ctx, serviceMap, prometheus.WOW, startTime, endTime, filter, granularity)

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
			latencyTempChartObject.ChartData = FillChart(startTime, endTime, stepMicros)
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
			errorTempChartObject.ChartData = FillChart(startTime, endTime, stepMicros)
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
			tpmTempChartObject.ChartData = FillChart(startTime, endTime, stepMicros)
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

func getStepMicros(duration int64) int64 {
	var stepMicros int64
	if duration <= 3_600_000_000 {
		stepMicros = 60_000_000 // 1m
	} else if duration <= 18_000_000_000 {
		stepMicros = 300_000_000 // 5m
	} else if duration <= 36_000_000_000 {
		stepMicros = 600_000_000 // 10m
	} else if duration <= 108_000_000_000 {
		stepMicros = 1_800_000_000 // 30m
	} else if duration <= 216_000_000_000 {
		stepMicros = 3_600_000_000 // 1h
	} else if duration <= 1_296_000_000_000 {
		stepMicros = 21_600_000_000 // 6h
	} else if duration <= 2_592_000_000_000 {
		stepMicros = 43_200_000_000 // 12h
	} else {
		stepMicros = 86_400_000_000 // 1d
	}
	return stepMicros
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

func FillChart(startTime, endTime time.Time, stepMicros int64) map[int64]float64 {
	values := make(map[int64]float64)
	for ts := startTime.UnixMicro(); ts <= endTime.UnixMicro(); ts += stepMicros {
		values[ts] = 0
	}
	return values
}

func GetRatio(value float64, base float64) float64 {
	ratio := float64(value / base)
	if base == 0 {
		return 0
	}

	if math.IsInf(ratio, 1) {
		return 9999999
	} else if math.IsInf(ratio, -1) {
		return -9999999
	} else {
		return (ratio - 1) * 100
	}
}
