package prometheus

import (
	"math"
)

// metricGroup Name
type MGroupName string

// metricName
type MName string

const (
	// metricGroup
	REALTIME MGroupName = "realtime" // endpoint时刻瞬时值
	AVG      MGroupName = "avg"      // start~endpoint之间的平均值
	DOD      MGroupName = "dod"      // start~endpoint时段和昨日日同比
	WOW      MGroupName = "wow"      // start~endpoint时段和上周周同比

	// metricName
	DEP_LATENCY MName = "dep_latency"
	LATENCY     MName = "latency"
	ERROR_RATE  MName = "error"
	THROUGHPUT  MName = "throughput"
)

type MetricGroupMap[K interface {
	comparable
	ConvertFromLabels
}, V MetricGroup] struct {
	// 用于返回列表
	MetricGroupList []V
	// 用于通过Key快速查询对应的MetricGroup
	MetricGroupMap map[K]V
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
				// 通常不会发生,这意味着initEmptyGroup返回的结构不是它本身
				continue
			}
			m.MetricGroupList = append(m.MetricGroupList, mg)
			m.MetricGroupMap[key] = mg
		}
		// 所有合并值均只包含最新时间点的结果,直接取metricResult.Values[0]
		value := metric.Values[0].Value
		if math.IsInf(value, 0) {
			continue
		}
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
