package service

import (
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"strings"
	"time"
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

// mergeMetrics 用于将指标结果合并到map中
// 功能上与MetricGroup.MergeMetricResults类似但是MergeMetricResults并不能完全复用
// TODO 修改MergeMetricResults，为所有指标提供setValue方法。使其能复用到所有结果的赋值
// TODO 无法合并指标labels中不含instanceKey的情况
func mergeMetrics(instances *InstanceMap, results []prometheus.MetricResult, metricName string) {
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
			instance.LatencyData = res.Values
		case metricErrorData:
			instance.ErrorRateData = res.Values
		case metricTPMData:
			// 计算结果为每秒次数，转为分钟
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

	filters = append(filters, prometheus.PodRegexPQLFilter, prometheus.ValueExistPQLValueFilter)
	filters = append(filters, prometheus.ContainerIdRegexPQLFilter, prometheus.ValueExistPQLValueFilter)
	return filters
}

// InstanceRED 获取instance粒度的RED指标
func (s *service) InstanceRED(startTime, endTime time.Time, filters []string) *InstanceMap {
	var res = &InstanceMap{
		MetricGroupList: []*prometheus.InstanceMetrics{},
		MetricGroupMap:  map[prometheus.InstanceKey]*prometheus.InstanceMetrics{},
	}

	s.promRepo.FillMetric(res, prometheus.AVG, startTime, endTime, filters, prometheus.InstanceGranularity)
	s.promRepo.FillMetric(res, prometheus.DOD, startTime, endTime, filters, prometheus.InstanceGranularity)
	s.promRepo.FillMetric(res, prometheus.WOW, startTime, endTime, filters, prometheus.InstanceGranularity)

	return res
}

// InstanceRangeData 获取instance粒度的RED指标的图表数据
func (s *service) InstanceRangeData(instances *InstanceMap, startTime, endTime time.Time, step time.Duration, filters []string) error {
	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	latencyRes, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgLatencyWithFilters,
		startTS,
		endTS,
		int64(step),
		prometheus.InstanceGranularity,
		filters...,
	)
	if err != nil {
		return err
	}
	mergeMetrics(instances, latencyRes, metricLatencyData)

	errorRes, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgErrorRateWithFilters,
		startTS,
		endTS,
		int64(step),
		prometheus.InstanceGranularity,
		filters...,
	)
	if err != nil {
		return err
	}
	mergeMetrics(instances, errorRes, metricErrorData)

	tpmRes, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgTPSWithFilters,
		startTS,
		endTS,
		int64(step),
		prometheus.InstanceGranularity,
		filters...,
	)
	if err != nil {
		return err
	}
	mergeMetrics(instances, tpmRes, metricTPMData)

	return nil
}

func mergeLogMetrics(instances *InstanceMap, results []prometheus.MetricResult, metricName string) {
	for _, res := range results {
		for key, value := range instances.MetricGroupMap {
			if key.Pod == res.Metric.POD {
				switch metricName {
				case metricLog:
					if &res.Values[0].Value != nil {
						value.LogAVGData = &res.Values[0].Value
					}
				case metricLogDOD:
					if &res.Values[0].Value != nil {
						value.LogDayOverDay = &res.Values[0].Value
					}
				case metricLogWOW:
					if &res.Values[0].Value != nil {
						value.LogWeekOverWeek = &res.Values[0].Value
					}
				case metricLogData:
					value.LogData = res.Values
				}
				break
			}
		}
	}
}

// InstanceLog 填充instance级别的log指标的均值、日同比、周同比、图表
func (s *service) InstanceLog(instances *InstanceMap, startTime, endTime time.Time, step time.Duration) error {
	startTS := startTime.UnixMicro()
	endTS := endTime.UnixMicro()

	var pods []string
	for key := range instances.MetricGroupMap {
		if len(key.Pod) > 0 {
			pods = append(pods, key.Pod)
		}
	}

	escapedKeys := make([]string, len(pods))
	for i, key := range pods {
		escapedKeys[i] = prometheus.EscapeRegexp(key)
	}
	// 使用 strings.Join 生成正则表达式模式
	regexPattern := strings.Join(escapedKeys, "|")
	filter := make([]string, 2)
	filter[0] = prometheus.LogMetricPodRegexPQLFilter
	filter[1] = regexPattern
	logDataRes, err := s.promRepo.QueryAggMetricsWithFilter(
		prometheus.PQLAvgLogErrorCountWithFilters,
		startTS,
		endTS,
		prometheus.LogGranularity,
		filter...)
	if err != nil {
		return err
	}
	mergeLogMetrics(instances, logDataRes, metricLog)

	logDODRes, err := s.promRepo.QueryAggMetricsWithFilter(
		prometheus.DayOnDay(prometheus.PQLAvgLogErrorCountWithFilters),
		startTS,
		endTS,
		prometheus.LogGranularity,
		filter...)
	if err != nil {
		return err
	}
	mergeLogMetrics(instances, logDODRes, metricLogDOD)

	logWOWRes, err := s.promRepo.QueryAggMetricsWithFilter(
		prometheus.WeekOnWeek(prometheus.PQLAvgLogErrorCountWithFilters),
		startTS,
		endTS,
		prometheus.LogGranularity,
		filter...)
	if err != nil {
		return err
	}
	mergeLogMetrics(instances, logWOWRes, metricLogWOW)

	logData, err := s.promRepo.QueryRangeAggMetricsWithFilter(
		prometheus.PQLAvgLogErrorCountWithFilters,
		startTS,
		endTS,
		int64(step),
		prometheus.LogGranularity,
		filter...,
	)
	if err != nil {
		return err
	}
	mergeLogMetrics(instances, logData, metricLogData)

	return nil
}
