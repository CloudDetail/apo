package service

import (
	"math"
	"sort"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	prom "github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/pkg/errors"
)

const (
	SortByLatency   = "latency"
	SortByErrorRate = "errorRate"
	SortByTps       = "tps"
)

func (s *service) GetSQLMetrics(req *request.GetSQLMetricsRequest) (*response.GetSQLMetricsResponse, error) {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)
	step := time.Duration(req.Step) * time.Microsecond

	sqlMetricMap := s.SQLREDMetric(startTime, endTime, req.Service)
	// 按平均延时/错误率/TPS 排序并分页
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

	// 分页
	var totalCount int
	if req.PageParam == nil {
		req.PageParam = &request.PageParam{
			CurrentPage: 1,
			PageSize:    99,
		}
	}
	sqlMetricMap.MetricGroupList, totalCount = pageByParam(sqlMetricMap.MetricGroupList, req.PageParam)
	var res = &response.GetSQLMetricsResponse{
		Pagination: model.Pagination{
			Total:       int64(totalCount),
			CurrentPage: req.PageParam.CurrentPage,
			PageSize:    req.PageParam.PageSize,
		},
		SQLOperationDetails: []response.SQLOperationDetail{},
	}

	// 填充chart
	_ = s.FillSQLREDChart(sqlMetricMap, req.Service, startTime, endTime, step)
	// 转换格式
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
			operation.LatencyChartData = convertToChart(avgLatency, true)
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
			operation.ErrorRateChartData = convertToChart(avgErrorRate, false)
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
			operation.TpsChartData = convertToChart(avgTPS, false)
		}
	}
	return nil
}

func convertToChart(result prom.MetricResult, transfromNano2Micro bool) map[int64]float64 {
	var data = make(map[int64]float64)
	for _, point := range result.Values {
		timestamp := point.TimeStamp
		value := point.Value
		if transfromNano2Micro {
			value = point.Value / 1e3
		}
		if !math.IsInf(value, 0) { //为无穷大时则不赋值
			data[timestamp] = value
		}
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
