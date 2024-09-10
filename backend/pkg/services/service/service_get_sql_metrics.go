package service

import (
	"math"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/pkg/errors"
)

func (s *service) GetSQLMetrics(req *request.GetSQLMetricsRequest) (*response.GetSQLMetricsResponse, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)
	step := time.Duration(req.Step) * time.Microsecond

	sqlMetricMap := s.SQLREDMetric(startTime, endTime, req.Service)

	err := s.FillSQLREDChart(sqlMetricMap, req.Service, startTime, endTime, step)
	if err != nil {

	}

	var res = &response.GetSQLMetricsResponse{
		SQLOperationDetails: []response.SQLOperationDetail{},
	}
	for _, metricGroups := range sqlMetricMap.MetricGroupList {
		// 清理空值
		metricGroups.REDMetrics.CleanUPNullValue()

		res.SQLOperationDetails = append(res.SQLOperationDetails, response.SQLOperationDetail{
			SQLKey: metricGroups.SQLKey,
			Latency: response.TempChartObject{
				ChartData: metricGroups.LatencyChartData,
				Value:     metricGroups.REDMetrics.Avg.Latency,
				Ratio: response.Ratio{
					DayOverDay:  metricGroups.REDMetrics.DOD.Latency,
					WeekOverDay: metricGroups.REDMetrics.WOW.Latency,
				},
			},
			ErrorRate: response.TempChartObject{
				ChartData: metricGroups.ErrorRateChartData,
				Value:     metricGroups.REDMetrics.Avg.ErrorRate,
				Ratio: response.Ratio{
					DayOverDay:  metricGroups.REDMetrics.DOD.ErrorRate,
					WeekOverDay: metricGroups.REDMetrics.WOW.ErrorRate,
				},
			},
			Tps: response.TempChartObject{
				ChartData: metricGroups.TpsChartData,
				Value:     metricGroups.REDMetrics.Avg.TPM,
				Ratio: response.Ratio{
					DayOverDay:  metricGroups.REDMetrics.DOD.TPM,
					WeekOverDay: metricGroups.REDMetrics.WOW.TPM,
				},
			},
		})
	}

	return res, nil
}

type SQLMetricMap = prom.MetricGroupMap[prom.SQLKey, *SQLMetricsWithChart]

type SQLMetricsWithChart struct {
	prom.SQLKey

	REDMetrics prom.REDMetrics

	LatencyChartData   map[int64]float64
	ErrorRateChartData map[int64]float64
	TpsChartData       map[int64]float64
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

// EndpointsREDMetric 查询Endpoint级别的RED指标结果(包括平均值,日同比变化率,周同比变化率)
func (s *service) SQLREDMetric(startTime, endTime time.Time, service string) *SQLMetricMap {
	var res = &SQLMetricMap{
		MetricGroupList: []*SQLMetricsWithChart{},
		MetricGroupMap:  map[prom.SQLKey]*SQLMetricsWithChart{},
	}

	var filters []string
	if len(service) > 0 {
		filters = append(filters, prom.ServicePQLFilter, service)
	}

	// 填充时间段内的平均RED指标
	s.fillSQLMetric(res, prom.AVG, startTime, endTime, filters)
	// 填充时间段内的RED指标日同比
	s.fillSQLMetric(res, prom.DOD, startTime, endTime, filters)
	// 填充时间段内的RED指标周同比
	s.fillSQLMetric(res, prom.WOW, startTime, endTime, filters)
	return res
}

func (s *service) fillSQLMetric(res *SQLMetricMap, metricGroup prom.MGroupName, startTime, endTime time.Time, filters []string) {
	// 装饰器,默认不修改PQL语句,用于AVG或REALTIME两个metricGroup
	var decorator = func(apf prom.AggPQLWithFilters) prom.AggPQLWithFilters {
		return apf
	}

	switch metricGroup {
	case prom.REALTIME:
		// 实时值使用当前时间往前3分钟作为时间间隔
		// 时间单位为microsecond
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
		// TODO 输出日志或记录错误到Endpoint中
	}

	res.MergeMetricResults(metricGroup, prom.LATENCY, latency)

	errorRate, err := s.promRepo.QueryAggMetricsWithFilter(
		decorator(prom.PQLAvgSQLErrorRateWithFilters),
		startTS, endTS,
		prom.DBOperationGranularity,
		filters...,
	)
	if err != nil {
		// TODO 输出日志或记录错误到Endpoint中
	}
	res.MergeMetricResults(metricGroup, prom.ERROR_RATE, errorRate)

	if metricGroup == prom.REALTIME {
		// 目前不计算TPS的实时值
		return
	}
	tps, err := s.promRepo.QueryAggMetricsWithFilter(
		decorator(prom.PQLAvgSQLTPSWithFilters),
		startTS, endTS,
		prom.DBOperationGranularity,
		filters...,
	)
	if err != nil {
		// TODO 输出日志或记录错误到Endpoint中
	}

	res.MergeMetricResults(metricGroup, prom.THROUGHPUT, tps)
}

// EndpointRangeREDChart 查询曲线图
func (s *service) FillSQLREDChart(sqlMap *SQLMetricMap, service string, startTime time.Time, endTime time.Time, step time.Duration) error {
	var opNames []string
	// 遍历 services 数组，获取每个 URL 的 ContentKey 并存储到切片中
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
			operation.LatencyChartData = convertToChart(avgLatency)
		}
	}

	avgErrorRates, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prom.PQLAvgSQLErrorRateWithFilters,
		startTime.UnixMicro(), endTime.UnixMicro(), step.Microseconds(),
		prom.DBOperationGranularity,
		filters...,
	)

	if err != nil {
		for _, avgErrorRate := range avgErrorRates {
			var sqlKey prom.SQLKey
			sqlKey = sqlKey.ConvertFromLabels(avgErrorRate.Metric).(prom.SQLKey)
			operation, find := sqlMap.MetricGroupMap[sqlKey]
			if !find {
				continue
			}
			operation.ErrorRateChartData = convertToChart(avgErrorRate)
		}
	}

	avgTPSs, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prom.PQLAvgSQLTPSWithFilters,
		startTime.UnixMicro(), endTime.UnixMicro(), step.Microseconds(),
		prom.DBOperationGranularity,
		filters...,
	)

	if err != nil {
		for _, avgTPS := range avgTPSs {
			var sqlKey prom.SQLKey
			sqlKey = sqlKey.ConvertFromLabels(avgTPS.Metric).(prom.SQLKey)
			operation, find := sqlMap.MetricGroupMap[sqlKey]
			if !find {
				continue
			}
			operation.TpsChartData = convertToChart(avgTPS)
		}
	}
	return nil
}

func convertToChart(result prom.MetricResult) map[int64]float64 {
	var data = make(map[int64]float64)
	for _, point := range result.Values {
		timestamp := point.TimeStamp
		value := point.Value
		if !math.IsInf(value, 0) { //为无穷大时则不赋值
			data[timestamp] = value
		}
	}
	return data
}
