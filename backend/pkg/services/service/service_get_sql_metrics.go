// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"sort"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/pkg/errors"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

const (
	SortByLatency	= "latency"
	SortByErrorRate	= "errorRate"
	SortByTps	= "tps"
)

func (s *service) GetSQLMetrics(ctx_core core.Context, req *request.GetSQLMetricsRequest) (*response.GetSQLMetricsResponse, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)
	step := time.Duration(req.Step) * time.Microsecond

	sqlMetricMap := s.SQLREDMetric(startTime, endTime, req.Service)
	// Sort and page by average latency/error rate/TPS
	switch req.SortBy {
	case SortByErrorRate:
		sort.Slice(sqlMetricMap.MetricGroupList, func(i, j int) bool {
			return compareValueWithNull(
				sqlMetricMap.MetricGroupList[i].REDMetrics.Avg.ErrorRate,
				sqlMetricMap.MetricGroupList[j].REDMetrics.Avg.ErrorRate)
		})
	case SortByTps:
		sort.Slice(sqlMetricMap.MetricGroupList, func(i, j int) bool {
			return compareValueWithNull(
				sqlMetricMap.MetricGroupList[i].REDMetrics.Avg.TPM,
				sqlMetricMap.MetricGroupList[j].REDMetrics.Avg.TPM)
		})
	default:
		sort.Slice(sqlMetricMap.MetricGroupList, func(i, j int) bool {
			return compareValueWithNull(
				sqlMetricMap.MetricGroupList[i].REDMetrics.Avg.Latency,
				sqlMetricMap.MetricGroupList[j].REDMetrics.Avg.Latency)
		})
	}

	// Paging
	var totalCount int
	if req.PageParam == nil {
		req.PageParam = &request.PageParam{
			CurrentPage:	1,
			PageSize:	99,
		}
	}
	sqlMetricMap.MetricGroupList, totalCount = pageByParam(sqlMetricMap.MetricGroupList, req.PageParam)
	var res = &response.GetSQLMetricsResponse{
		Pagination: model.Pagination{
			Total:		int64(totalCount),
			CurrentPage:	req.PageParam.CurrentPage,
			PageSize:	req.PageParam.PageSize,
		},
		SQLOperationDetails:	[]response.SQLOperationDetail{},
	}

	// Fill the chart
	_ = s.FillSQLREDChart(sqlMetricMap, req.Service, startTime, endTime, step)
	// Convert format
	for _, metricGroups := range sqlMetricMap.MetricGroupList {
		res.SQLOperationDetails = append(res.SQLOperationDetails, response.SQLOperationDetail{
			SQLKey:	metricGroups.SQLKey,
			Latency: response.TempChartObject{
				ChartData:	metricGroups.LatencyChartData,
				Value:		metricGroups.REDMetrics.Avg.Latency,
				Ratio: response.Ratio{
					DayOverDay:	metricGroups.REDMetrics.DOD.Latency,
					WeekOverDay:	metricGroups.REDMetrics.WOW.Latency,
				},
			},
			ErrorRate: response.TempChartObject{
				ChartData:	metricGroups.ErrorRateChartData,
				Value:		metricGroups.REDMetrics.Avg.ErrorRate,
				Ratio: response.Ratio{
					DayOverDay:	metricGroups.REDMetrics.DOD.ErrorRate,
					WeekOverDay:	metricGroups.REDMetrics.WOW.ErrorRate,
				},
			},
			Tps: response.TempChartObject{
				ChartData:	metricGroups.TpsChartData,
				Value:		metricGroups.REDMetrics.Avg.TPM,
				Ratio: response.Ratio{
					DayOverDay:	metricGroups.REDMetrics.DOD.TPM,
					WeekOverDay:	metricGroups.REDMetrics.WOW.TPM,
				},
			},
		})
	}

	return res, nil
}

func compareValueWithNull(valueI *float64, valueJ *float64) bool {
	if valueI == nil && valueJ == nil {
		return true
	} else if valueI == nil {
		return false
	} else if valueJ == nil {
		return true
	}
	return *valueI > *valueJ
}

type SQLMetricMap = prom.MetricGroupMap[prom.SQLKey, *SQLMetricsWithChart]

type SQLMetricsWithChart struct {
	prom.SQLKey

	REDMetrics	prom.REDMetrics

	LatencyChartData	map[int64]float64
	ErrorRateChartData	map[int64]float64
	TpsChartData		map[int64]float64
}

func (s *SQLMetricsWithChart) InitEmptyGroup(key prom.ConvertFromLabels) prom.MetricGroup {
	return &SQLMetricsWithChart{
		SQLKey: key.(prom.SQLKey),
	}
}

func (s *SQLMetricsWithChart) AppendGroupIfNotExist(metricGroup prom.MGroupName, metricName prom.MName) bool {
	return metricName == prom.LATENCY
}

func (s *SQLMetricsWithChart) SetValue(metricGroup prom.MGroupName, metricName prom.MName, value float64) {
	s.REDMetrics.SetValue(metricGroup, metricName, value)
}

func (s *SQLMetricsWithChart) SetValues(metricGroup prom.MGroupName, metricName prom.MName, points []prom.Points) {
	s.REDMetrics.SetValues(metricGroup, metricName, points)
}

// EndpointsREDMetric query Endpoint-level RED metric results (including average value, DoD/WoW Growth Rate)
func (s *service) SQLREDMetric(startTime, endTime time.Time, service string) *SQLMetricMap {
	var res = &SQLMetricMap{
		MetricGroupList:	[]*SQLMetricsWithChart{},
		MetricGroupMap:		map[prom.SQLKey]*SQLMetricsWithChart{},
	}

	var filters []string
	if len(service) > 0 {
		filters = append(filters, prom.ServicePQLFilter, service)
	}

	// Average RED metric over the fill time period
	s.fillSQLMetric(res, prom.AVG, startTime, endTime, filters)
	// RED metric day-to-day-on-da during the fill period
	s.fillSQLMetric(res, prom.DOD, startTime, endTime, filters)
	// RED metric week-on-week in the fill time period
	s.fillSQLMetric(res, prom.WOW, startTime, endTime, filters)
	return res
}

func (s *service) fillSQLMetric(res *SQLMetricMap, metricGroup prom.MGroupName, startTime, endTime time.Time, filters []string) {
	// Decorator, PQL statement is not modified by default, for AVG or REALTIME two metricGroup
	var decorator = func(apf prom.AggPQLWithFilters) prom.AggPQLWithFilters {
		return apf
	}

	switch metricGroup {
	case prom.REALTIME:
		// real-time value uses 3 minutes ahead of current time as time interval
		// Time unit is microsecond
		startTime = endTime.Add(-3 * time.Minute)
	case prom.DOD:
		decorator = prom.DayOnDay
	case prom.WOW:
		decorator = prom.WeekOnWeek
	}

	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	latency, err := s.promRepo.QueryAggMetricsWithFilter(
		decorator(prom.PQLAvgSQLLatencyWithFilters),
		startTS, endTS,
		prom.DBOperationGranularity,
		filters...,
	)
	if err != nil {
		// TODO output log or log errors to Endpoint
	}

	res.MergeMetricResults(metricGroup, prom.LATENCY, latency)

	errorRate, err := s.promRepo.QueryAggMetricsWithFilter(
		decorator(prom.PQLAvgSQLErrorRateWithFilters),
		startTS, endTS,
		prom.DBOperationGranularity,
		filters...,
	)
	if err != nil {
		// TODO output log or log errors to Endpoint
	}
	res.MergeMetricResults(metricGroup, prom.ERROR_RATE, errorRate)

	if metricGroup == prom.REALTIME {
		// Currently, the real-time value of TPS is not calculated.
		return
	}
	tps, err := s.promRepo.QueryAggMetricsWithFilter(
		decorator(prom.PQLAvgSQLTPSWithFilters),
		startTS, endTS,
		prom.DBOperationGranularity,
		filters...,
	)
	if err != nil {
		// TODO output log or log errors to Endpoint
	}

	res.MergeMetricResults(metricGroup, prom.THROUGHPUT, tps)
}

// EndpointRangeREDChart query graph
func (s *service) FillSQLREDChart(sqlMap *SQLMetricMap, service string, startTime time.Time, endTime time.Time, step time.Duration) error {
	var opNames []string
	// Traverse the services array, get the ContentKey of each URL and store it in the slice.
	for _, sqlOperation := range sqlMap.MetricGroupList {
		opNames = append(opNames, sqlOperation.DBOperation)
		service = sqlOperation.Service
	}

	var filters []string
	if len(service) > 0 {
		filters = append(filters, prom.ServicePQLFilter, service)
	}
	if len(opNames) > 0 {
		filters = append(filters, prom.DBNameRegexPQLFilter, prom.RegexMultipleValue(opNames...))
	} else {
		return errors.New("no sql operation found")
	}

	avgLatencys, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prom.PQLAvgSQLLatencyWithFilters,
		startTime.UnixMicro(), endTime.UnixMicro(), step.Microseconds(),
		prom.DBOperationGranularity,
		filters...,
	)
	if err == nil {
		for _, avgLatency := range avgLatencys {
			var sqlKey prom.SQLKey
			sqlKey = sqlKey.ConvertFromLabels(avgLatency.Metric).(prom.SQLKey)
			operation, find := sqlMap.MetricGroupMap[sqlKey]
			if !find {
				continue
			}
			operation.LatencyChartData = convertToChart(avgLatency, prom.AVG, prom.LATENCY)
		}
	}

	avgErrorRates, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prom.PQLAvgSQLErrorRateWithFilters,
		startTime.UnixMicro(), endTime.UnixMicro(), step.Microseconds(),
		prom.DBOperationGranularity,
		filters...,
	)

	if err == nil {
		for _, avgErrorRate := range avgErrorRates {
			var sqlKey prom.SQLKey
			sqlKey = sqlKey.ConvertFromLabels(avgErrorRate.Metric).(prom.SQLKey)
			operation, find := sqlMap.MetricGroupMap[sqlKey]
			if !find {
				continue
			}
			operation.ErrorRateChartData = convertToChart(avgErrorRate, prom.AVG, prom.ERROR_RATE)
		}
	}

	avgTPSs, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prom.PQLAvgSQLTPSWithFilters,
		startTime.UnixMicro(), endTime.UnixMicro(), step.Microseconds(),
		prom.DBOperationGranularity,
		filters...,
	)

	if err == nil {
		for _, avgTPS := range avgTPSs {
			var sqlKey prom.SQLKey
			sqlKey = sqlKey.ConvertFromLabels(avgTPS.Metric).(prom.SQLKey)
			operation, find := sqlMap.MetricGroupMap[sqlKey]
			if !find {
				continue
			}
			operation.TpsChartData = convertToChart(avgTPS, prom.AVG, prom.THROUGHPUT)
		}
	}
	return nil
}

func convertToChart(result prom.MetricResult, metricGroup prom.MGroupName, metricName prom.MName) map[int64]float64 {
	var data = make(map[int64]float64)
	for _, point := range result.Values {
		adjustValue := prom.AdjustREDValue(metricGroup, metricName, point.Value)
		data[point.TimeStamp] = adjustValue
	}
	return data
}

func pageByParam(list []*SQLMetricsWithChart, param *request.PageParam) ([]*SQLMetricsWithChart, int) {
	totalCount := len(list)
	if param == nil {
		return list, totalCount
	}

	if totalCount < param.PageSize {
		return list, totalCount
	}

	startIdx := (param.CurrentPage - 1) * param.PageSize
	endIdx := startIdx + param.PageSize
	if endIdx > totalCount {
		endIdx = totalCount
	}
	return list[startIdx:endIdx], totalCount
}
