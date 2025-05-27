// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"fmt"
	"math"
	"time"

	"github.com/hashicorp/go-multierror"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

const (
	metricLatencyData = "latencyData"
	metricErrorData   = "errorData"
	metricTPMData     = "TPMData"
	metricLog         = "log"
	metricLogDOD      = "logDOD"
	metricLogWOW      = "logWOW"
	metricLogData     = "logData" // chart
)

type InstanceMap = prometheus.MetricGroupMap[prometheus.InstanceKey, *prometheus.InstanceMetrics]

func IsInvalidData(m map[prometheus.InstanceKey]*prometheus.InstanceMetrics, metric *prometheus.InstanceMetrics) bool {
	if len(metric.Pod) == 0 && len(metric.PID) == 0 && len(metric.ContainerId) == 0 {
		return true
	}

	if metric.PID == "1" {
		// skip process with the same information except pid = 1
		for k := range m {
			if k.Pod == metric.Pod &&
				k.ContainerId == metric.ContainerId &&
				k.Namespace == metric.Namespace &&
				k.NodeName == metric.NodeName &&
				k.NodeIP == metric.NodeIP {
				return true
			}
		}
	}

	return false
}

// mergeChartMetrics is used to merge metric results into map
// Functionally similar to MetricGroup.MergeMetricResults but MergeMetricResults not fully reusable
// TODO modify the MergeMetricResults to provide setValue methods for all metrics. Assignment to enable reuse to all results
// TODO cannot merge the situation that the metric labels do not contain instanceKey
func mergeChartMetrics(instances *InstanceMap, results []prometheus.MetricResult, metricName string) {
	for _, res := range results {
		var kType prometheus.InstanceKey
		key := kType.ConvertFromLabels(res.Metric).(prometheus.InstanceKey)

		instance, ok := instances.MetricGroupMap[key]
		// The delay information obtained before is no longer filled because there is no delay information corresponding to the instance.
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

// InstanceRED get the RED metric of instance granularity
func (s *service) InstanceRED(ctx core.Context, startTime, endTime time.Time, filters []string, res *InstanceMap) {
	s.promRepo.FillMetric(ctx, res, prometheus.AVG, startTime, endTime, filters, prometheus.InstanceGranularity)
	s.promRepo.FillMetric(ctx, res, prometheus.DOD, startTime, endTime, filters, prometheus.InstanceGranularity)
	s.promRepo.FillMetric(ctx, res, prometheus.WOW, startTime, endTime, filters, prometheus.InstanceGranularity)

}

// InstanceRangeData get chart data for the RED metric with instance granularity
func (s *service) InstanceRangeData(ctx core.Context, instances *InstanceMap, startTime, endTime time.Time, step time.Duration, filters []string) *multierror.Error {
	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	latencyRes, latencyErr := s.promRepo.QueryRangeAggMetricsWithFilter(ctx,
		prometheus.PQLAvgLatencyWithFilters,
		startTS,
		endTS,
		step.Microseconds(),
		prometheus.InstanceGranularity,
		filters...,
	)
	mergeChartMetrics(instances, latencyRes, metricLatencyData)

	errorRes, rateErr := s.promRepo.QueryRangeAggMetricsWithFilter(ctx,
		prometheus.PQLAvgErrorRateWithFilters,
		startTS,
		endTS,
		step.Microseconds(),
		prometheus.InstanceGranularity,
		filters...,
	)
	mergeChartMetrics(instances, errorRes, metricErrorData)

	tpmRes, tmpErr := s.promRepo.QueryRangeAggMetricsWithFilter(ctx,
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

// InstanceLog fill the mean value, DoD/WoW Growth Rate, and chart of log metrics filled with instance levels.
func (s *service) InstanceLog(ctx core.Context, instances *InstanceMap, startTime, endTime time.Time, step time.Duration) *multierror.Error {
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

	logDataRes, avgErr := s.promRepo.QueryData(ctx, endTime, pql)
	mergeLogMetrics(instances, logDataRes, metricLog)

	pqlDOD := `(` + pql + `) / ((` + pql + `) offset 24h )`
	logDODRes, dodErr := s.promRepo.QueryData(ctx, endTime, pqlDOD)
	mergeLogMetrics(instances, logDODRes, metricLogDOD)

	pqlWOW := `(` + pql + `) / ((` + pql + `) offset 7d )`
	logWOWRes, wowErr := s.promRepo.QueryData(ctx, endTime, pqlWOW)
	mergeLogMetrics(instances, logWOWRes, metricLogWOW)

	logData, chartErr := s.promRepo.QueryInstanceLogRangeData(
		ctx,
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

func (s *service) GetNormalLog(ctx core.Context, startTime, endTime time.Time, filterKVs []string, offset time.Duration) []prometheus.MetricResult {
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
	data, _ := s.promRepo.QueryData(ctx, endTime.Add(-offset), pql)
	return data
}
