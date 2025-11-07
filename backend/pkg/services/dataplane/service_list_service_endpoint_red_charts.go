package dataplane

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (s *service) ListServiceRedCharts(ctx core.Context, req *request.ListServiceRedChartsRequest) *response.ListServiceRedChartsResponse {
	req.StartTime = util.RoundDownToMinute(req.StartTime)
	stepMicros := getStepMicros(req.EndTime - req.StartTime)
	startBucket := req.StartTime / stepMicros
	endBucket := req.EndTime / stepMicros

	resp := make(map[string][]*response.QueryChartResult)

	// results := make([]*response.QueryChartResult, 0)
	bucketMetrics, err := s.chRepo.ListGroupServiceRedMetrics(ctx, req.StartTime, req.EndTime, req.Cluster, stepMicros)
	if err != nil {
		return &response.ListServiceRedChartsResponse{
			Msg: "query redmetrics: " + err.Error(),
		}
	}
	if len(bucketMetrics) == 0 {
		return &response.ListServiceRedChartsResponse{
			Results: resp,
		}
	}

	serviceMap := make(map[string]map[int64]*clickhouse.ServiceBucketRedMetric)

	for _, bucketMetric := range bucketMetrics {
		bucketMap, find := serviceMap[bucketMetric.ServiceName]
		if !find {
			bucketMap = make(map[int64]*clickhouse.ServiceBucketRedMetric)
		}
		bucketMap[int64(bucketMetric.TimeBucket)] = &bucketMetric
		serviceMap[bucketMetric.ServiceName] = bucketMap
	}

	for service, bucketMap := range serviceMap {
		results := make([]*response.QueryChartResult, 0)

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

		// dodMetrics, err := s.chRepo.QueryGroupServiceRedMetricValue(ctx, req.StartTime-24*int64(time.Hour)/1000, req.StartTime, req.Cluster, service, "")
		// if err != nil {
		// 	return &response.ListServiceRedChartsResponse{
		// 		Msg: "query dod metric failed: " + err.Error(),
		// 	}
		// }
		// wowMetrics, err := s.chRepo.QueryGroupServiceRedMetricValue(ctx, req.StartTime-7*24*int64(time.Hour)/1000, req.StartTime, req.Cluster, service, "")
		// if err != nil {
		// 	return &response.ListServiceRedChartsResponse{
		// 		Msg: "query wow metric failed: " + err.Error(),
		// 	}
		// }
		avgLatency := groupMetrics.GetAvgLatency()
		results = append(results, &response.QueryChartResult{
			Title: "Response Time",
			Unit:  "us",
			Timeseries: []*response.Timeseries{
				{
					Legend:       service,
					LegendFormat: "",
					Labels: map[string]string{
						"service": service,
					},
					Chart: response.TempChartObject{
						ChartData: latencyChart,
						Ratio:     response.Ratio{
							// DayOverDay:  util.PtrFloat64(GetRatio(avgLatency, dodMetrics.GetAvgLatency())),
							// WeekOverDay: util.PtrFloat64(GetRatio(avgLatency, wowMetrics.GetAvgLatency())),
						},
						Value: util.PtrFloat64(avgLatency),
					},
				},
			},
		})
		errorRate := groupMetrics.GetErrorRate()
		results = append(results, &response.QueryChartResult{
			Title: "Error Rate",
			Unit:  "%",
			Timeseries: []*response.Timeseries{
				{
					Legend:       service,
					LegendFormat: "",
					Labels: map[string]string{
						"service": service,
					},
					Chart: response.TempChartObject{
						ChartData: errorChart,
						Ratio:     response.Ratio{
							// DayOverDay:  util.PtrFloat64(GetRatio(errorRate, dodMetrics.GetErrorRate())),
							// WeekOverDay: util.PtrFloat64(GetRatio(errorRate, wowMetrics.GetErrorRate())),
						},
						Value: util.PtrFloat64(errorRate),
					},
				},
			},
		})
		tpm := groupMetrics.GetTpm()
		results = append(results, &response.QueryChartResult{
			Title: "Tpm",
			Unit:  "requests/minute",
			Timeseries: []*response.Timeseries{
				{
					Legend:       service,
					LegendFormat: "",
					Labels: map[string]string{
						"service": service,
					},
					Chart: response.TempChartObject{
						ChartData: countChart,
						Ratio:     response.Ratio{
							// DayOverDay:  util.PtrFloat64(GetRatio(tpm, dodMetrics.GetTpm())),
							// WeekOverDay: util.PtrFloat64(GetRatio(tpm, wowMetrics.GetTpm())),
						},
						Value: util.PtrFloat64(tpm),
					},
				},
			},
		})

		resp[service] = results
	}
	return &response.ListServiceRedChartsResponse{
		Results: resp,
	}
}
