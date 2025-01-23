// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

// metricGroup Name
type MGroupName string

// metricName
type MName string

const (
	// metricGroup
	REALTIME MGroupName = "realtime"
	AVG      MGroupName = "avg"
	DOD      MGroupName = "dod" // Day-over-Day Growth Rate
	WOW      MGroupName = "wow" // Week-over-Week Growth Rate

	// metricName
	DEP_LATENCY     MName = "dep_latency"
	LATENCY         MName = "latency"
	ERROR_RATE      MName = "error"
	THROUGHPUT      MName = "throughput"
	LOG_ERROR_COUNT MName = "log_error_count"
)

type MetricGroupMap[K interface {
	comparable
	ConvertFromLabels
}, V MetricGroup] struct {
	// used to return a list
	MetricGroupList []V
	// Used to quickly query the corresponding key by MetricGroup
	MetricGroupMap map[K]V
}

type MetricGroupInterface interface {
	MergeMetricResults(metricGroup MGroupName, metricName MName, metricResults []MetricResult)
}

func (m *MetricGroupMap[K, V]) MergeMetricResults(metricGroup MGroupName, metricName MName, metricResults []MetricResult) {
	for _, metric := range metricResults {
		if len(metric.Values) <= 0 {
			continue
		}
		var kType K
		key, ok := kType.ConvertFromLabels(metric.Metric).(K)
		if !ok {
			continue
		}
		mg, find := m.MetricGroupMap[key]
		var pMG = new(V)
		if !find {
			if !(*pMG).AppendGroupIfNotExist(metricGroup, metricName) {
				continue
			}
			mg, ok = mg.InitEmptyGroup(key).(V)
			if !ok {
				continue
			}
			m.MetricGroupList = append(m.MetricGroupList, mg)
			m.MetricGroupMap[key] = mg
		}
		// All consolidated values contain only the results at the latest time point, directly take the metricResult.Values[0]
		value := metric.Values[0].Value
		mg.SetValue(metricGroup, metricName, value)
	}
}

type ConvertFromLabels interface {
	ConvertFromLabels(labels Labels) ConvertFromLabels
}

type MetricGroup interface {
	InitEmptyGroup(key ConvertFromLabels) MetricGroup
	AppendGroupIfNotExist(metricGroup MGroupName, metricName MName) bool
	SetValue(metricGroup MGroupName, metricName MName, value float64)
}
