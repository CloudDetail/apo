// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"math"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

const (
	metricLatencyData = "latencyData"
	metricErrorData   = "errorData"
	metricTPMData     = "TPMData"
	metricLog         = "log"
	metricLogDOD      = "logDOD"
	metricLogWOW      = "logWOW"
	metricLogData     = "logData" // 图表
)

type InstanceMap = prometheus.MetricGroupMap[prometheus.InstanceKey, *prometheus.InstanceMetrics]

// mergeChartMetrics 用于将指标结果合并到map中
// 功能上与MetricGroup.MergeMetricResults类似但是MergeMetricResults并不能完全复用
// TODO 修改MergeMetricResults，为所有指标提供setValue方法。使其能复用到所有结果的赋值
// TODO 无法合并指标labels中不含instanceKey的情况
func mergeChartMetrics(instances *InstanceMap, results []prometheus.MetricResult, metricName string) {
	for _, res := range results {
		var kType prometheus.InstanceKey
		key := kType.ConvertFromLabels(res.Metric).(prometheus.InstanceKey)

		instance, ok := instances.MetricGroupMap[key]
		// 之前获取过延时信息，这里没有对应instance的延时信息故不再填充
		if !ok {
			continue
		}
		switch metricName {
		case metricLatencyData:
			for i := range res.Values {
				res.Values[i].Value /= 1e3
			}
			instance.LatencyData = res.Values
		case metricErrorData:
			for i := range res.Values {
				res.Values[i].Value *= 100
			}
			instance.ErrorRateData = res.Values
		case metricTPMData:
			for i := range res.Values {
				res.Values[i].Value *= 60
			}
			instance.TPMData = res.Values
		}
	}
}

type InstancesFilter struct {
	SrvName    string
	ContentKey string
}

func (f InstancesFilter) ExtractFilterStr() []string {
	var filters []string
	if len(f.SrvName) > 0 {
		filters = append(filters, prometheus.ServicePQLFilter, f.SrvName)
	}
	if len(f.ContentKey) > 0 {
		filters = append(filters, prometheus.ContentKeyPQLFilter, f.ContentKey)
	}

	filters = append(filters, prometheus.PodRegexPQLFilter, prometheus.LabelExistPQLValueFilter)
	filters = append(filters, prometheus.ContainerIdRegexPQLFilter, prometheus.LabelExistPQLValueFilter)
	return filters
}

// InstanceRED 获取instance粒度的RED指标
func (s *service) InstanceRED(startTime, endTime time.Time, filters []string, res *InstanceMap) {
	s.promRepo.FillMetric(res, prometheus.AVG, startTime, endTime, filters, prometheus.InstanceGranularity)
	s.promRepo.FillMetric(res, prometheus.DOD, startTime, endTime, filters, prometheus.InstanceGranularity)
	s.promRepo.FillMetric(res, prometheus.WOW, startTime, endTime, filters, prometheus.InstanceGranularity)

}

// InstanceRangeData 获取instance粒度的RED指标的图表数据
func (s *service) InstanceRangeData(instances *InstanceMap, startTime, endTime time.Time, step time.Duration, filters []string) *multierror.Error {
	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	latencyRes, latencyErr := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgLatencyWithFilters,
		startTS,
		endTS,
		step.Microseconds(),
		prometheus.InstanceGranularity,
		filters...,
	)
	mergeChartMetrics(instances, latencyRes, metricLatencyData)

	errorRes, rateErr := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgErrorRateWithFilters,
		startTS,
		endTS,
		step.Microseconds(),
		prometheus.InstanceGranularity,
		filters...,
	)
	mergeChartMetrics(instances, errorRes, metricErrorData)

	tpmRes, tmpErr := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgTPSWithFilters,
		startTS,
		endTS,
		step.Microseconds(),
		prometheus.InstanceGranularity,
		filters...,
	)
	mergeChartMetrics(instances, tpmRes, metricTPMData)

	var err *multierror.Error
	err = multierror.Append(err, latencyErr, rateErr, tmpErr)
	return err
}

func adjustValue(value *float64) {
	if math.IsInf(*value, 1) {
		*value = prometheus.RES_MAX_VALUE
	} else if math.IsInf(*value, -1) {
		*value = -prometheus.RES_MAX_VALUE
	} else {
		*value = (*value - 1) * 100
	}
}

func mergeLogMetrics(instances *InstanceMap, results []prometheus.MetricResult, metricName string) {
	for _, res := range results {
		for key, value := range instances.MetricGroupMap {
			if key.Pod == res.Metric.POD || (key.PID == res.Metric.PID && key.NodeName == res.Metric.NodeName) {
				switch metricName {
				case metricLog:
					value.LogAVGData = &res.Values[0].Value
				case metricLogDOD:
					adjustValue(&res.Values[0].Value)
					value.LogDayOverDay = &res.Values[0].Value
				case metricLogWOW:
					adjustValue(&res.Values[0].Value)
					value.LogWeekOverWeek = &res.Values[0].Value
				case metricLogData:
					value.LogData = res.Values
				}
				break
			}
		}
	}
}

// InstanceLog 填充instance级别的log指标的均值、日同比、周同比、图表
func (s *service) InstanceLog(instances *InstanceMap, startTime, endTime time.Time, step time.Duration) *multierror.Error {
	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	var pods, nodeName, pid []string
	for key := range instances.MetricGroupMap {
		if len(key.Pod) > 0 {
			pods = append(pods, key.Pod)
		}
		if len(key.NodeName) > 0 {
			nodeName = append(nodeName, key.NodeName)
		}
		if len(key.PID) > 0 {
			pid = append(pid)
		}
	}

	podFilter := make([]string, 2)
	podFilter[0] = prometheus.LogMetricPodRegexPQLFilter
	podFilter[1] = prometheus.RegexMultipleValue(pods...)
	vmFilter := make([]string, 4)
	vmFilter[0] = prometheus.LogMetricNodeRegexPQLFilter
	vmFilter[1] = prometheus.RegexMultipleValue(nodeName...)
	vmFilter[2] = prometheus.LogMetricPidRegexPQLFilter
	vmFilter[3] = prometheus.RegexMultipleValue(pid...)
	pql, pqlErr := prometheus.PQLInstanceLog(
		prometheus.PQLAvgLogErrorCountWithFilters,
		startTS, endTS,
		prometheus.LogGranularity,
		podFilter, vmFilter)

	var err *multierror.Error
	if pqlErr != nil {
		err = multierror.Append(err, pqlErr)
		return err
	}

	logDataRes, avgErr := s.promRepo.QueryData(endTime, pql)
	mergeLogMetrics(instances, logDataRes, metricLog)

	pqlDOD := `(` + pql + `) / ((` + pql + `) offset 24h )`
	logDODRes, dodErr := s.promRepo.QueryData(endTime, pqlDOD)
	mergeLogMetrics(instances, logDODRes, metricLogDOD)

	pqlWOW := `(` + pql + `) / ((` + pql + `) offset 7d )`
	logWOWRes, wowErr := s.promRepo.QueryData(endTime, pqlWOW)
	mergeLogMetrics(instances, logWOWRes, metricLogWOW)

	logData, chartErr := s.promRepo.QueryInstanceLogRangeData(
		prometheus.PQLAvgLogErrorCountWithFilters,
		startTS,
		endTS,
		step.Microseconds(),
		prometheus.LogGranularity,
		podFilter, vmFilter,
	)
	mergeLogMetrics(instances, logData, metricLogData)

	err = multierror.Append(err, avgErr, dodErr, wowErr, chartErr)
	return err
}

func (s *service) GetNormalLog(startTime, endTime time.Time, filterKVs []string, offset time.Duration) []prometheus.MetricResult {
	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	if len(filterKVs)%2 != 0 {
		return nil
	}
	var filters []string
	for i := 0; i+1 < len(filterKVs); i += 2 {
		filters = append(filters, fmt.Sprintf("%s\"%s\"", filterKVs[i], filterKVs[i+1]))
	}
	vector := prometheus.VecFromS2E(startTS, endTS)
	pql := prometheus.PQLNormalLogCountWithFilters(vector, string(prometheus.LogGranularity), filters)
	data, _ := s.promRepo.QueryData(endTime.Add(-offset), pql)
	return data
}
